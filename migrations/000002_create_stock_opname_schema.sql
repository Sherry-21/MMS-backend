-- 1. Table: Stock Opnames (Cycle count / physical inventory audit sessions)
CREATE TABLE IF NOT EXISTS stock_opnames (
    id BIGSERIAL PRIMARY KEY,
    opname_number VARCHAR(100) NOT NULL UNIQUE,
    opname_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    total_items INT NOT NULL DEFAULT 0,
    total_discrepancy INT NOT NULL DEFAULT 0,
    notes TEXT NULL,
    assigned_by VARCHAR(36) NULL,
    conducted_by VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL
);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_date ON stock_opnames (opname_date);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_status ON stock_opnames (status);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_number ON stock_opnames (opname_number);

-- 2. Table: Stock Opname Items (Line items for each audited product variant)
CREATE TABLE IF NOT EXISTS stock_opname_items (
    id BIGSERIAL PRIMARY KEY,
    stock_opname_id BIGINT NOT NULL,
    product_variant_id BIGINT NOT NULL,
    system_quantity INT NOT NULL DEFAULT 0,
    physical_quantity INT NULL,
    discrepancy INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stock_opname_id) REFERENCES stock_opnames(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_opname_items_opname ON stock_opname_items (stock_opname_id);
CREATE INDEX IF NOT EXISTS idx_opname_items_variant ON stock_opname_items (product_variant_id);
CREATE INDEX IF NOT EXISTS idx_opname_items_status ON stock_opname_items (status);
