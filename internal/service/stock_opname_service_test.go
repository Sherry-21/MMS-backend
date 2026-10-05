package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"maiora-backend-service/internal/model"
	"maiora-backend-service/internal/payloads"
	"maiora-backend-service/internal/repository"
)

type mockStockOpnameRepo struct {
	opnames       map[int64]*model.StockOpname
	mockVariants  []model.ProductVariantWithStock
	nextID        int64
	countForDates map[string]int64
}

func newMockStockOpnameRepo() *mockStockOpnameRepo {
	return &mockStockOpnameRepo{
		opnames:       make(map[int64]*model.StockOpname),
		countForDates: make(map[string]int64),
		nextID:        1,
	}
}

func (m *mockStockOpnameRepo) CreateOpnameWithItems(ctx context.Context, opname *model.StockOpname, items []model.StockOpnameItem) error {
	opname.ID = m.nextID
	m.nextID++
	opname.TotalItems = len(items)

	for i := range items {
		items[i].ID = int64(i + 1)
		items[i].StockOpnameID = opname.ID
		// Attach mock product variant
		items[i].ProductVariant = &model.ProductVariant{
			ID:          items[i].ProductVariantID,
			SKU:         "SKU-TEST",
			VariantName: "Variant Test",
			Product: &model.Product{
				Name: "Product Test",
			},
		}
	}
	opname.Items = items
	m.opnames[opname.ID] = opname
	return nil
}

func (m *mockStockOpnameRepo) GetOpnameByID(ctx context.Context, id int64) (*model.StockOpname, error) {
	op, ok := m.opnames[id]
	if !ok {
		return nil, repository.ErrStockOpnameNotFound
	}
	return op, nil
}

func (m *mockStockOpnameRepo) GetTodayActiveOpname(ctx context.Context, today time.Time) (*model.StockOpname, error) {
	dateStr := today.Format("2006-01-02")
	for _, op := range m.opnames {
		if op.OpnameDate.Format("2006-01-02") == dateStr && (op.Status == model.OpnameStatusPending || op.Status == model.OpnameStatusInProgress) {
			return op, nil
		}
	}
	return nil, nil
}

func (m *mockStockOpnameRepo) ListOpnames(ctx context.Context, status string, fromDate, toDate string, limit, offset int) ([]model.StockOpname, int64, error) {
	var list []model.StockOpname
	for _, op := range m.opnames {
		if status != "" && op.Status != status {
			continue
		}
		list = append(list, *op)
	}
	return list, int64(len(list)), nil
}

func (m *mockStockOpnameRepo) UpdateOpnameItemCounts(ctx context.Context, opnameID int64, updates []payloads.CountItemInput) error {
	op, ok := m.opnames[opnameID]
	if !ok {
		return repository.ErrStockOpnameNotFound
	}

	for _, u := range updates {
		found := false
		for i := range op.Items {
			if op.Items[i].ID == u.ItemID {
				found = true
				qty := u.PhysicalQuantity
				op.Items[i].PhysicalQuantity = &qty
				op.Items[i].Discrepancy = qty - op.Items[i].SystemQuantity
				op.Items[i].Status = model.OpnameItemStatusCounted
				op.Items[i].Notes = u.Notes
			}
		}
		if !found {
			return repository.ErrOpnameItemNotFound
		}
	}
	op.Status = model.OpnameStatusInProgress
	return nil
}

func (m *mockStockOpnameRepo) CompleteOpname(ctx context.Context, opnameID int64, conductedBy string, notes string) (*model.StockOpname, error) {
	op, ok := m.opnames[opnameID]
	if !ok {
		return nil, repository.ErrStockOpnameNotFound
	}

	totalDiscrepancy := 0
	for _, item := range op.Items {
		totalDiscrepancy += item.Discrepancy
	}

	now := time.Now().UTC()
	op.Status = model.OpnameStatusCompleted
	op.TotalDiscrepancy = totalDiscrepancy
	op.CompletedAt = &now
	op.ConductedBy = &conductedBy
	op.Notes = notes

	return op, nil
}

func (m *mockStockOpnameRepo) SelectRandomVariantsForOpname(ctx context.Context, sampleSize int) ([]model.ProductVariantWithStock, error) {
	if len(m.mockVariants) == 0 {
		return nil, nil
	}
	if sampleSize > len(m.mockVariants) {
		sampleSize = len(m.mockVariants)
	}
	return m.mockVariants[:sampleSize], nil
}

func (m *mockStockOpnameRepo) CountOpnamesForDate(ctx context.Context, date time.Time) (int64, error) {
	return m.countForDates[date.Format("2006-01-02")], nil
}

func TestStockOpnameService_GenerateDailyAssignment_Success(t *testing.T) {
	repo := newMockStockOpnameRepo()
	repo.mockVariants = []model.ProductVariantWithStock{
		{ProductVariantID: 1, SKU: "SKU-1", VariantName: "128GB", ProductName: "iPhone 15", CurrentStock: 50},
		{ProductVariantID: 2, SKU: "SKU-2", VariantName: "256GB", ProductName: "iPhone 15", CurrentStock: 30},
	}
	svc := NewStockOpnameService(repo)

	assignedBy := "user-admin-1"
	resp, err := svc.GenerateDailyAssignment(context.Background(), &payloads.GenerateStockOpnameRequest{
		SampleSize: 2,
		Notes:      "Morning audit",
	}, &assignedBy)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.TotalItems != 2 {
		t.Errorf("expected 2 items, got %d", resp.TotalItems)
	}
	if resp.Status != model.OpnameStatusPending {
		t.Errorf("expected status PENDING, got %s", resp.Status)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 line items in response, got %d", len(resp.Items))
	}
	if resp.Items[0].SystemQuantity != 50 {
		t.Errorf("expected system quantity 50, got %d", resp.Items[0].SystemQuantity)
	}
}

func TestStockOpnameService_GenerateDailyAssignment_ReturnsExistingActive(t *testing.T) {
	repo := newMockStockOpnameRepo()
	repo.mockVariants = []model.ProductVariantWithStock{
		{ProductVariantID: 1, SKU: "SKU-1", CurrentStock: 50},
	}
	svc := NewStockOpnameService(repo)

	// First generation
	first, err := svc.GenerateDailyAssignment(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("failed first generation: %v", err)
	}

	// Second generation call on same day should return the same session
	second, err := svc.GenerateDailyAssignment(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("failed second generation: %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("expected same session ID %d, got %d", first.ID, second.ID)
	}
}

func TestStockOpnameService_SubmitItemCounts_And_Complete(t *testing.T) {
	repo := newMockStockOpnameRepo()
	repo.mockVariants = []model.ProductVariantWithStock{
		{ProductVariantID: 1, SKU: "SKU-1", CurrentStock: 50},
		{ProductVariantID: 2, SKU: "SKU-2", CurrentStock: 20},
	}
	svc := NewStockOpnameService(repo)

	opname, err := svc.GenerateDailyAssignment(context.Background(), &payloads.GenerateStockOpnameRequest{SampleSize: 2}, nil)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Submit physical count: item 1 had 50 system, counted 48 (-2 discrepancy)
	// item 2 had 20 system, counted 21 (+1 discrepancy)
	countReq := &payloads.SubmitStockOpnameCountRequest{
		Items: []payloads.CountItemInput{
			{ItemID: opname.Items[0].ID, PhysicalQuantity: 48, Notes: "2 broken"},
			{ItemID: opname.Items[1].ID, PhysicalQuantity: 21, Notes: "1 extra found"},
		},
	}

	counted, err := svc.SubmitItemCounts(context.Background(), opname.ID, countReq)
	if err != nil {
		t.Fatalf("submit counts failed: %v", err)
	}

	if counted.Status != model.OpnameStatusInProgress {
		t.Errorf("expected status IN_PROGRESS, got %s", counted.Status)
	}
	if counted.Items[0].Discrepancy != -2 {
		t.Errorf("expected discrepancy -2, got %d", counted.Items[0].Discrepancy)
	}
	if counted.Items[1].Discrepancy != 1 {
		t.Errorf("expected discrepancy 1, got %d", counted.Items[1].Discrepancy)
	}

	// Finalize / Complete Opname
	completeReq := &payloads.CompleteStockOpnameRequest{
		ConductedBy: "Staff John",
		Notes:       "All counted and verified",
	}
	final, err := svc.CompleteOpname(context.Background(), opname.ID, completeReq)
	if err != nil {
		t.Fatalf("complete opname failed: %v", err)
	}

	if final.Status != model.OpnameStatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", final.Status)
	}
	// Total discrepancy = -2 + 1 = -1
	if final.TotalDiscrepancy != -1 {
		t.Errorf("expected total discrepancy -1, got %d", final.TotalDiscrepancy)
	}
	if final.CompletedAt == nil {
		t.Errorf("expected completed_at to be populated")
	}
}

func TestStockOpnameService_NoVariantsError(t *testing.T) {
	repo := newMockStockOpnameRepo()
	repo.mockVariants = []model.ProductVariantWithStock{} // empty
	svc := NewStockOpnameService(repo)

	_, err := svc.GenerateDailyAssignment(context.Background(), nil, nil)
	if !errors.Is(err, repository.ErrNoVariantsAvailable) {
		t.Errorf("expected ErrNoVariantsAvailable, got %v", err)
	}
}
