# Maiora System Backend

A robust, clean-architecture backend service built in **Go (Golang)** using the **Chi** HTTP router, adhering to the **Model - Service - Repository** pattern with interfaces, environment configuration management, and centralized utilities.

---

## 📂 Folder Structure

```
maiora-backend/
├── cmd/
│   ├── api/
│   │   └── main.go                     # Entry point for HTTP API server
│   └── migrate/
│       └── main.go                     # Standalone CLI tool for database migrations
├── config/
│   └── config.go                       # Application, Server, Database & CORS configuration
├── docs/                               # Auto-generated Swagger / OpenAPI spec files
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/                           # Private application business code
│   ├── handler/                        # HTTP / Transport layer (Chi routing & handlers)
│   │   ├── auth_handler.go             # Authentication HTTP handler with Swagger annotations
│   │   ├── auth_handler_test.go        # HTTP endpoint integration tests
│   │   ├── import_batch_handler.go     # Import Batch HTTP handler with Swagger annotations
│   │   ├── import_batch_handler_test.go# HTTP endpoint integration tests
│   │   └── router.go                   # Chi router, middlewares, and route registration
│   ├── model/                          # Domain entities (Pure Database Models)
│   │   ├── category.go                 # Category GORM database entity
│   │   ├── product.go                  # Product GORM database entity
│   │   ├── product_variant.go          # ProductVariant GORM database entity
│   │   ├── import_batch.go             # ImportBatch GORM database entity
│   │   ├── import_batch_item.go        # ImportBatchItem GORM database entity
│   │   ├── inventory_batch.go          # InventoryBatch GORM database entity
│   │   └── user.go                     # User GORM database entity
│   ├── payloads/                       # Request & Response DTOs
│   │   ├── import_batch.go             # CreateImportBatchRequest, ImportBatchResponse
│   │   ├── user.go                     # RegisterRequest, LoginRequest, UserResponse, LoginResponse
│   │   └── response.go                 # Generic API response envelope (BaseResponse)
│   ├── repository/                     # Data Access layer (Interface-driven, GORM backed)
│   │   └── user_repository.go          # UserRepository interface & GORM implementation
│   └── service/                        # Business Logic layer (Interface-driven)
│       ├── auth_service.go             # AuthService interface & implementation
│       └── auth_service_test.go        # Unit tests with mock repositories
├── pkg/                                # Public reusable libraries & utilities
│   ├── database/                       # Database connection & migration helpers
│   │   └── mysql.go                    # GORM MySQL connection pool & AutoMigrate
│   └── utils/                          # Shared Go utility helpers
│       ├── env.go                      # Env reader with typed fallbacks
│       ├── env_test.go                 # Unit tests for env utils
│       ├── password.go                 # Bcrypt password hashing & validation
│       └── response.go                 # Standardized JSON response helpers
├── .env.example                        # Template for environment configuration
├── .env                                # Local environment file
├── .gitignore                          # Git ignore rules
├── go.mod                              # Go module definition
└── go.sum                              # Go dependency checksums
```

---

## ⚙️ Database Configuration & Migration Tool

### 1. Database Credentials in `.env`
```env
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASSWORD="password"
DB_NAME="maiora"
```

### 2. Running Database Migrations (`cmd/migrate`)
To create or update database tables:

```bash
go run ./cmd/migrate
```

---

## 🚀 Running the API Server (`cmd/api`)

```bash
go run ./cmd/api
```
The server will start listening on `http://localhost:8080`.

---

## 📡 API Endpoints & Swagger

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/` | Root status endpoint |
| `GET` | `/health` | Service health check |
| `POST` | `/api/v1/auth/register` | **Register a new user** (bcrypt hashed) |
| `POST` | `/api/v1/auth/login` | **User login** (credential verification) |
| `GET` | `/swagger/*` | **Swagger Interactive API Documentation UI** |
The server will start listening on `http://localhost:8080`.
The server will start listening on `http://localhost:8080`.

### 4. Run Tests
```bash
go test -v ./...
```

---

| `POST` | `/api/v1/categories` | **Create product category** (with slug) |
| `GET` | `/api/v1/categories` | **List all categories** |
| `POST` | `/api/v1/products` | **Create product** (with optional inline variants) |
| `GET` | `/api/v1/products` | **List products** (with category & variants) |
| `GET` | `/api/v1/products/{id}` | **Get product by ID** |
| `POST` | `/api/v1/products/{id}/variants` | **Add a variant/SKU to a product** |
| `GET` | `/api/v1/product-variants` | **List all variants/SKUs** (for Import Batch dropdowns) |
| `POST` | `/api/v1/import-batches` | **Register an import batch** (Atomic transaction + auto-populates inventory) |
| `GET` | `/swagger/*` | **Swagger Interactive API Documentation UI** |

### 📖 Accessing Swagger UI
When the server is running, open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

---

## 🏬 Product Center & Import Batch Workflow Example

### 1. Create a Category
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Smartphones", "slug": "smartphones"}'
```

### 2. Create a Product with Initial Variants/SKUs
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
        "variant_name": "Natural Titanium 128GB",
        "default_selling_price": "18999000.00"
      },
      {
        "sku": "IP15PRO-BLK-256",
        "variant_name": "Black Titanium 256GB",
        "default_selling_price": "21999000.00"
      }
    ]
  }'
```

### 3. List All Product Variants (For Search / Dropdown)
```bash
curl -X GET "http://localhost:8080/api/v1/product-variants?search=IP15PRO"
```

### 4. Register an Import Batch (Referencing Product Variant IDs)
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
      }
    ]
  }'
```

---

## 📦 Example: Registering an Import Batch

```bash
curl -X POST http://localhost:8080/api/v1/import-batches \
  -H "Content-Type: application/json" \
  -d '{
    "batch_number": "IMP-2026-09-001",
    "import_date": "2026-09-20T00:00:00Z",
    "supplier_name": "Shenzhen Global Tech Ltd",
    "currency_code": "USD",
    "exchange_rate": "15850.5000",
    "items": [
      {
        "product_variant_id": 1,
        "quantity_received": 100,
        "unit_cost_original_currency": "12.5000"
      },
      {
        "product_variant_id": 2,
        "quantity_received": 50,
        "unit_cost_original_currency": "25.0000"
      }
    ]
  }'
```

---

## 🔒 Security & Best Practices Implemented

1. **User ID Primary Key**: Uses UUID string format (`column:user_id;type:varchar(36);primaryKey`).
2. **Password Hashing**: Uses **`bcrypt`** (`golang.org/x/crypto/bcrypt`) with work factor `12`.
3. **Information Disclosure Prevention**: Password hashes are tagged with `json:"-"` and never returned in API payloads.
4. **Data Transfer Objects (DTOs)**: `RegisterRequest`, `LoginRequest`, `UserResponse`, `LoginResponse` keep database models decoupled from presentation.

---

## 🚀 Running the Services

### 1. Apply Database Migrations (Creates `users` & `greetings` tables)
```bash
go run ./cmd/migrate
```

### 2. Start the API Server
```bash
go run ./cmd/api
```

### Example Response: `GET /api/v1/hello`
```json
{
  "success": true,
  "message": "Greeting fetched successfully",
  "data": {
    "message": "Hello, World! Welcome to Maiora System Backend",
    "timestamp": "2026-08-15T08:34:53.123456Z"
  },
  "timestamp": "2026-08-15T08:34:53.123456Z"
}
```

---

## 🏗️ Architectural Flow

1. **HTTP Request** hits Chi router (`internal/handler/router.go`) through configured middlewares (RequestID, RealIP, Logger, Recoverer, Timeout, CORS).
2. Router forwards request to **Handler** (`internal/handler/greeting_handler.go`).
3. Handler parses query/body and invokes **Service Interface** (`internal/service/greeting_service.go`).
4. Service executes business logic and communicates with **Repository Interface** (`internal/repository/greeting_repository.go`).
5. Repository fetches data from database/storage and returns domain **Model** (`internal/model/greeting.go`).
6. Handler formats the result via **Utils** (`pkg/utils/response.go`) into a standardized JSON response.
