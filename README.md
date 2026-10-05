# Maiora System Backend (MMS-backend)

A robust, clean-architecture backend service built in **Go (Golang)** using the **Chi** HTTP router, adhering to the **Model - Service - Repository** pattern with interfaces, environment configuration management, standardized response envelopes, and centralized utilities.

---

## 📂 Folder Structure

```
maiora-backend/
├── cmd/
│   ├── api/
│   │   └── main.go                         # Entry point for HTTP API server
│   └── migrate/
│       └── main.go                         # Standalone CLI tool for database migrations
├── config/
│   └── config.go                           # Application, Server, Database & CORS configuration
├── docs/                                   # Auto-generated Swagger / OpenAPI spec files
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/                               # Private application business code
│   ├── handler/                            # HTTP / Transport layer (Chi routing & handlers)
│   │   ├── auth_handler.go                 # Authentication HTTP handler
│   │   ├── auth_handler_test.go            # Authentication endpoint integration tests
│   │   ├── category_handler.go             # Category HTTP handler
│   │   ├── import_batch_handler.go         # Import batch HTTP handler
│   │   ├── import_batch_handler_test.go    # Import batch integration tests
│   │   ├── product_handler.go              # Product & SKU variant HTTP handler
│   │   ├── product_handler_test.go         # Product & response format integration tests
│   │   ├── stock_opname_handler.go         # Stock opname & cycle count HTTP handler
│   │   ├── stock_opname_handler_test.go    # Stock opname handler integration tests
│   │   └── router.go                       # Chi router setup, route groups & endpoint mappings
│   ├── middleware/                         # HTTP Middlewares (Transport & domain filters)
│   │   ├── auth_middleware.go              # Stateless JWT authentication & user verification
│   │   └── auth_middleware_test.go         # AuthMiddleware unit tests
│   ├── model/                              # Domain entities (GORM database models)
│   │   ├── category.go                     # Category entity
│   │   ├── product.go                      # Product catalog entity
│   │   ├── product_variant.go              # Product SKU variant entity
│   │   ├── import_batch.go                 # Import batch header entity
│   │   ├── import_batch_item.go            # Import batch line item entity
│   │   ├── inventory_batch.go              # Inventory batch tracking entity
│   │   ├── refresh_token.go                # Refresh token entity (id, user_id, refresh_token, expires_at, is_revoked)
│   │   ├── stock_opname.go                 # Stock opname & item audit entities
│   │   └── user.go                         # User entity (UUID primary key & bcrypt auth)
│   ├── payloads/                           # Request & Response Data Transfer Objects (DTOs)
│   │   ├── category.go                     # Category request/response DTOs
│   │   ├── import_batch.go                 # Import batch request/response DTOs
│   │   ├── product.go                      # Product & variant request/response DTOs
│   │   ├── response.go                     # Standard BaseResponse & PaginationMeta DTOs
│   │   ├── response_test.go                # BaseResponse & pagination serialization tests
│   │   ├── stock_opname.go                 # Stock opname request/response DTOs
│   │   └── user.go                         # Auth & user request/response DTOs
│   ├── repository/                         # Data Access layer (Interface-driven, GORM backed)
│   │   ├── category_repository.go          # Category database operations
│   │   ├── import_batch_repository.go      # Import batch & inventory transactions
│   │   ├── product_repository.go           # Product & variant catalog queries
│   │   ├── refresh_token_repository.go     # Refresh token persistence & revocation queries
│   │   ├── stock_opname_repository.go      # Stock opname & random sampling queries
│   │   └── user_repository.go              # User database operations
│   └── service/                            # Business Logic layer (Interface-driven)
│       ├── auth_service.go                 # User registration & password verification
│       ├── auth_service_test.go            # AuthService unit tests with mock repository
│       ├── category_service.go             # Category creation & slug handling
│       ├── import_batch_service.go         # Import batch calculation & inventory creation
│       ├── import_batch_service_test.go    # ImportBatchService unit tests
│       ├── product_service.go              # Product & variant catalog logic
│       ├── product_service_test.go         # ProductService unit tests
│       ├── stock_opname_service.go         # Daily random cycle count & audit logic
│       └── stock_opname_service_test.go    # StockOpnameService unit tests
├── migrations/                             # Raw SQL schema definition scripts
│   ├── 000001_create_inventory_and_import_schema.sql
│   ├── 000002_create_stock_opname_schema.sql
│   └── 000003_create_refresh_tokens_table.sql
├── pkg/                                    # Public reusable libraries & utilities
│   ├── database/                           # Database connection & migration helpers
│   │   └── postgres.go                     # GORM PostgreSQL connection pool & AutoMigrate (Neon DB ready)
│   └── utils/                              # Shared Go utility helpers
│       ├── env.go                          # Environment reader with typed fallbacks
│       ├── env_test.go                     # Unit tests for env utils
│       ├── jwt.go                          # JWT generation, token pair, and claims validation
│       ├── jwt_test.go                     # Unit tests for JWT signing & expiration
│       ├── password.go                     # Bcrypt password hashing & validation
│       └── response.go                     # Standardized JSON response helpers
├── .env.example                            # Template for environment configuration
├── .env                                    # Local environment file
├── .gitignore                              # Git ignore rules
├── go.mod                                  # Go module definition
└── go.sum                                  # Go dependency checksums
```

---

## ⚙️ Environment Configuration

Copy `.env.example` to `.env` and set your credentials:

```env
# Application
APP_NAME="Maiora System Backend"
APP_ENV="development"
PORT="8080"

# Database Configuration (Neon DB / PostgreSQL)
# Option 1: Direct Neon Connection URI (Recommended)
DATABASE_URL="postgres://user:password@ep-sample-123456.us-east-2.aws.neon.tech/neondb?sslmode=require"

# Option 2: Discrete DB Parameters
DB_HOST="ep-sample-123456.us-east-2.aws.neon.tech"
DB_PORT="5432"
DB_USER="user"
DB_PASSWORD="password"
DB_NAME="neondb"
DB_SSLMODE="require"

# CORS
CORS_ALLOWED_ORIGINS="*"
CORS_ALLOWED_METHODS="GET,POST,PUT,DELETE,OPTIONS"
CORS_ALLOWED_HEADERS="Accept,Authorization,Content-Type,X-CSRF-Token"

# JWT Authentication
JWT_SECRET="maiora-super-secret-jwt-key-change-in-production"
JWT_ACCESS_EXPIRATION_MINUTES=15
JWT_REFRESH_EXPIRATION_DAYS=7
```

---

## 🗄️ Database Migrations

Run database migrations to initialize tables in Neon DB (`users`, `categories`, `products`, `product_variants`, `import_batches`, `import_batch_items`, `inventory_batches`, `stock_opnames`, `stock_opname_items`, `refresh_tokens`):

```bash
go run ./cmd/migrate
```

---

## 🚀 Running the API Server

Start the HTTP server on port `8080` (or configured `PORT`):

```bash
go run ./cmd/api
```

---

## 🐳 Running with Docker & Docker Compose

Run automatic migrations and the MMS backend API connected to Neon DB:

```bash
# Build and start services (migrate & api)
docker compose up --build -d

# Check service logs
docker compose logs -f api

# Stop containers
docker compose down
```

To run the Docker image standalone connecting to Neon DB:

```bash
# Build image
docker build -t maiora-backend .

# Run container with Neon DB connection string
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@ep-sample-123456.us-east-2.aws.neon.tech/neondb?sslmode=require" \
  maiora-backend
```

---

## 🧪 Running Automated Tests

Run all unit and integration test suites:

```bash
go test -v ./...
```

---

## 📋 Standard API Response Specification

Every endpoint returns a unified JSON envelope following this structure:

```json
{
  "success": true,
  "message": "Human readable message",
  "meta": {},
  "data": ...
}
```

### 1. Non-Paginated Responses (`GET` by ID, `GET` all, `POST`, etc.)
When pagination is not applicable, `meta` defaults to an empty JSON object `{}`:

```json
{
  "success": true,
  "message": "Category created successfully",
  "meta": {},
  "data": {
    "id": 1,
    "name": "Smartphones",
    "slug": "smartphones",
    "created_at": "2026-09-20T14:00:00Z"
  }
}
```

### 2. Paginated Responses (`GET /api/v1/products`)
For paginated endpoints, `meta` contains pagination metadata, and `data` contains the item collection directly:

```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  },
  "data": [
    {
      "id": 1,
      "category_id": 1,
      "name": "iPhone 15 Pro",
      "description": "Titanium design, A17 Pro chip",
      "category": {
        "id": 1,
        "name": "Smartphones",
        "slug": "smartphones"
      },
      "variants": [
        {
          "id": 1,
          "product_id": 1,
          "sku": "IP15PRO-NAT-128",
          "variant_name": "Natural Titanium 128GB",
          "created_at": "2026-09-20T14:00:00Z"
        }
      ],
      "created_at": "2026-09-20T14:00:00Z"
    }
  ]
}
```

### 3. Error Responses
```json
{
  "success": false,
  "message": "Invalid request payload",
  "meta": {},
  "error": "sku is required"
}
```

---

## 📡 API Endpoints & Swagger

### Interactive Swagger UI
When the server is running, navigate to:
```
http://localhost:8080/swagger/index.html
```

To regenerate Swagger documentation after modifying handler annotations or models:
```bash
swag init -g cmd/api/main.go
```

### Route Summary

| Method | Endpoint | Description | Auth Required | Pagination |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/` | Root server status check | No | `meta: {}` |
| `GET` | `/health` | Service health check | No | `meta: {}` |
| `POST` | `/api/v1/auth/register` | Register a new user (`bcrypt` hashed) | No | `meta: {}` |
| `POST` | `/api/v1/auth/login` | Authenticate user, return JWT access + refresh tokens | No | `meta: {}` |
| `POST` | `/api/v1/auth/refresh` | Rotate tokens using active refresh token | No | `meta: {}` |
| `GET` | `/api/v1/auth/me` | Get current user (validates token & session match) | **Bearer JWT** | `meta: {}` |
| `POST` | `/api/v1/categories` | Create a product category with slug | No | `meta: {}` |
| `GET` | `/api/v1/categories` | List all categories alphabetically | No | `meta: {}` |
| `POST` | `/api/v1/products` | Create a product with optional inline variants | No | `meta: {}` |
| `GET` | `/api/v1/products` | List products (`search`, `category_id`, `page`, `limit`) | No | **Paginated** |
| `GET` | `/api/v1/products/{id}` | Get product details with category and variants | No | `meta: {}` |
| `POST` | `/api/v1/products/{id}/variants` | Add a SKU variant to an existing product | No | `meta: {}` |
| `POST` | `/api/v1/import-batches` | Register import batch & populate inventory batches | No | `meta: {}` |
| `POST` | `/api/v1/stock-opnames/generate-daily` | Generate or fetch today's random cycle count assignment | Optional | `meta: {}` |
| `GET` | `/api/v1/stock-opnames/today` | Get active cycle count assigned for today | No | `meta: {}` |
| `GET` | `/api/v1/stock-opnames` | List stock opname history (`page`, `limit`, `status`, `from_date`, `to_date`) | No | **Paginated** |
| `GET` | `/api/v1/stock-opnames/{id}` | Get full stock opname session detail and line items | No | `meta: {}` |
| `POST` | `/api/v1/stock-opnames/{id}/count` | Submit counted physical quantities & notes | No | `meta: {}` |
| `POST` | `/api/v1/stock-opnames/{id}/complete` | Finalize stock opname session and compute total discrepancy | Optional | `meta: {}` |

---

## 🏬 Product Catalog & Import Workflow Examples

### 1. Create a Product Category
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Smartphones", "slug": "smartphones"}'
```

### 2. Create a Product with Variants
> **Note:** Product variants only contain SKU identification (`sku`, `variant_name`).
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": 1,
    "name": "iPhone 15 Pro",
    "description": "Titanium design, A17 Pro chip",
    "variants": [
      {
        "sku": "IP15PRO-NAT-128",
        "variant_name": "Natural Titanium 128GB"
      },
      {
        "sku": "IP15PRO-BLK-256",
        "variant_name": "Black Titanium 256GB"
      }
    ]
  }'
```

### 3. Add an Additional Variant to a Product
```bash
curl -X POST http://localhost:8080/api/v1/products/1/variants \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "IP15PRO-WHT-512",
    "variant_name": "White Titanium 512GB"
  }'
```

### 4. Search Product Variants for Import Batch Selection
```bash
curl -X GET "http://localhost:8080/api/v1/product-variants?search=IP15PRO"
```

### 5. Register an Import Batch (Multi-currency & Auto-Inventory)
```bash
curl -X POST http://localhost:8080/api/v1/import-batches \
  -H "Content-Type: application/json" \
  -d '{
    "batch_number": "IMP-2026-09-001",
    "import_date": "2026-09-20T00:00:00Z",
    "supplier_name": "Apple Distribution International",
    "currency_code": "USD",
    "exchange_rate": "15850.5000",
    "items": [
      {
        "product_variant_id": 1,
        "quantity_received": 100,
        "unit_cost_original_currency": "850.0000"
      },
      {
        "product_variant_id": 2,
        "quantity_received": 50,
        "unit_cost_original_currency": "950.0000"
      }
    ]
  }'
```

---

## 🔍 Stock Opname (Daily Random Cycle Count) Workflow

Stock Opname enables warehouse staff to audit inventory physically against system records.

### 1. Generate Daily Random Assignment
Randomly samples a set of product variants (e.g. 5 items) and snapshots current on-hand stock:
```bash
curl -X POST http://localhost:8080/api/v1/stock-opnames/generate-daily \
  -H "Content-Type: application/json" \
  -d '{"sample_size": 5, "notes": "Morning routine cycle count"}'
```

### 2. View Today's Active Assignment
```bash
curl -X GET http://localhost:8080/api/v1/stock-opnames/today
```

### 3. Record Physical Counts & Notes
Staff records physical counts; discrepancies (`physical_quantity - system_quantity`) are calculated automatically:
```bash
curl -X POST http://localhost:8080/api/v1/stock-opnames/1/count \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "item_id": 1,
        "physical_quantity": 98,
        "notes": "2 units missing from shelf"
      },
      {
        "item_id": 2,
        "physical_quantity": 50,
        "notes": "Exact match"
      }
    ]
  }'
```

### 4. Finalize & Complete Stock Opname
Computes total discrepancy across all items, changes status to `COMPLETED`, and timestamps completion:
```bash
curl -X POST http://localhost:8080/api/v1/stock-opnames/1/complete \
  -H "Content-Type: application/json" \
  -d '{
    "conducted_by": "Staff John",
    "notes": "Verified and signed off"
  }'
```

### 5. Browse Historical Records
Filter by date range, completion status, and paginate through past audit sessions:
```bash
curl -X GET "http://localhost:8080/api/v1/stock-opnames?status=COMPLETED&from_date=2026-09-01&page=1&limit=20"
```

---

## 🔒 Security & Best Practices

1. **UUID Primary Keys**: Users utilize UUID v4 strings for opaque primary key identification.
2. **Password Security**: Bcrypt with work factor `12`. Hashes are hidden (`json:"-"`) from responses.
3. **Transactional Integrity**: Import batch registration atomically creates headers, items, and inventory batches in a single DB transaction.
4. **Clean Architecture**: Strong boundary isolation via interfaces between Handlers (HTTP), Services (Business Logic), and Repositories (Data Access).
5. **Stateless Access Tokens**: Access tokens are signed JWTs verified statelessly via signature and expiration. Access tokens are not stored in the database.
6. **Dedicated `refresh_tokens` Table & Rotation (RTR)**: Refresh tokens are persisted in a dedicated 5-column table (`id`, `user_id`, `refresh_token`, `expires_at`, `is_revoked`). On every refresh request, the previous token is marked `is_revoked = true` and a new token pair is issued, effectively preventing token replay and reuse attacks.
