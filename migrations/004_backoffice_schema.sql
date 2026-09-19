-- Migration 004: Backoffice schema for multi-branch, profit sharing, payroll
PRAGMA foreign_keys = ON;

-- ============================================================
-- Branches
-- ============================================================
CREATE TABLE IF NOT EXISTS branches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  address TEXT NOT NULL DEFAULT '',
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO branches (code, name, address) VALUES ('KLASEMAN', 'Pardis Barbershop Klaseman', 'Klaseman');
INSERT OR IGNORE INTO branches (code, name, address) VALUES ('LEDOK', 'Pardis Barbershop Ledok', 'Ledok');

-- ============================================================
-- Catalog Items (services & products with commission tracking)
-- ============================================================
CREATE TABLE IF NOT EXISTS catalog_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  item_type TEXT NOT NULL CHECK (item_type IN ('SERVICE', 'PRODUCT')) DEFAULT 'SERVICE',
  price_cents INTEGER NOT NULL DEFAULT 0,
  commission_amount INTEGER NOT NULL DEFAULT 0,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- Branch Profit Sharing Rules (per-branch, per-month config)
-- ============================================================
CREATE TABLE IF NOT EXISTS branch_profit_sharing_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  branch_id INTEGER NOT NULL REFERENCES branches(id),
  period_month TEXT NOT NULL,
  owner_percentage REAL NOT NULL DEFAULT 0,
  unallocated_percentage REAL NOT NULL DEFAULT 0,
  updated_by INTEGER REFERENCES users(id),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(branch_id, period_month)
);

-- ============================================================
-- Employee Profit Sharing Rules (per-employee allocation)
-- ============================================================
CREATE TABLE IF NOT EXISTS employee_profit_sharing_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  rule_id INTEGER NOT NULL REFERENCES branch_profit_sharing_rules(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users(id),
  percentage REAL NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(rule_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_employee_profit_sharing_rule ON employee_profit_sharing_rules(rule_id);

-- ============================================================
-- Discounts and Bundles
-- ============================================================
CREATE TABLE IF NOT EXISTS discounts_and_bundles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('PERCENTAGE', 'FIXED_AMOUNT', 'BUNDLE')),
  value INTEGER NOT NULL DEFAULT 0,
  service_allocation_ratio REAL NOT NULL DEFAULT 1.0,
  product_allocation_ratio REAL NOT NULL DEFAULT 0.0,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
