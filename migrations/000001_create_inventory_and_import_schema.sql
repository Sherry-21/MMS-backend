-- 1. Table: Categories
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories (slug);

-- 2. Table: Products
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);

-- 3. Table: Product Variants
CREATE TABLE IF NOT EXISTS product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    sku VARCHAR(100) NOT NULL UNIQUE,
    variant_name VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_product_variants_sku ON product_variants (sku);
CREATE INDEX IF NOT EXISTS idx_product_variants_product ON product_variants (product_id);

-- 4. Table: Import Batches
CREATE TABLE IF NOT EXISTS import_batches (
    id BIGSERIAL PRIMARY KEY,
    batch_number VARCHAR(100) NOT NULL UNIQUE,
    import_date DATE NOT NULL,
    supplier_name VARCHAR(255) NOT NULL,
    currency_code VARCHAR(10) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(18, 4) NOT NULL,
    total_import_cost DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_import_batches_number ON import_batches (batch_number);
CREATE INDEX IF NOT EXISTS idx_import_batches_date ON import_batches (import_date);

-- 5. Table: Import Batch Items
CREATE TABLE IF NOT EXISTS import_batch_items (
    id BIGSERIAL PRIMARY KEY,
    import_batch_id BIGINT NOT NULL,
    product_variant_id BIGINT NOT NULL,
    quantity_received INT NOT NULL CHECK (quantity_received > 0),
    unit_cost_original_currency DECIMAL(18, 4) NOT NULL,
    unit_cost_idr DECIMAL(18, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (import_batch_id) REFERENCES import_batches(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_import_items_batch ON import_batch_items (import_batch_id);
CREATE INDEX IF NOT EXISTS idx_import_items_variant ON import_batch_items (product_variant_id);

-- 6. Table: Inventory Batches
CREATE TABLE IF NOT EXISTS inventory_batches (
    id BIGSERIAL PRIMARY KEY,
    import_batch_item_id BIGINT NOT NULL,
    product_variant_id BIGINT NOT NULL,
    initial_quantity INT NOT NULL CHECK (initial_quantity >= 0),
    remaining_quantity INT NOT NULL CHECK (remaining_quantity >= 0),
    unit_cost_idr DECIMAL(18, 2) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (import_batch_item_id) REFERENCES import_batch_items(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_inventory_variant_remaining ON inventory_batches (product_variant_id, remaining_quantity);
CREATE INDEX IF NOT EXISTS idx_inventory_item ON inventory_batches (import_batch_item_id);

-- 7. Table: Users
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
