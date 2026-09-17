PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT,username TEXT UNIQUE,email TEXT NOT NULL UNIQUE,display_name TEXT NOT NULL,password_hash TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN ('superadmin','operator','barberman','cashier','manager')),active INTEGER NOT NULL DEFAULT 1,branch_id INTEGER REFERENCES branches(id),staff_type TEXT CHECK(staff_type IN ('barberman','cashier','manager','owner') OR staff_type IS NULL),phone_number TEXT,bank_name TEXT,bank_account_number TEXT,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE TABLE IF NOT EXISTS sessions (token_hash TEXT PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,expires_at DATETIME NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expires_at);
CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY AUTOINCREMENT,operator_id INTEGER NOT NULL REFERENCES users(id),kind TEXT NOT NULL CHECK(kind IN ('income','expense')),amount_cents INTEGER NOT NULL CHECK(amount_cents > 0),category TEXT NOT NULL,note TEXT NOT NULL DEFAULT '',occurred_at DATETIME NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,reversal_of_id INTEGER REFERENCES transactions(id),branch_id INTEGER REFERENCES branches(id));
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_transactions_operator_date ON transactions(operator_id,occurred_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_one_reversal ON transactions(reversal_of_id) WHERE reversal_of_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_branch ON transactions(branch_id);

CREATE TABLE IF NOT EXISTS audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_id INTEGER NOT NULL REFERENCES users(id),
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  details TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_events_created ON audit_events(created_at DESC, id DESC);

CREATE TRIGGER IF NOT EXISTS transactions_immutable_update
BEFORE UPDATE ON transactions BEGIN
  SELECT RAISE(ABORT, 'transactions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS transactions_immutable_delete
BEFORE DELETE ON transactions BEGIN
  SELECT RAISE(ABORT, 'transactions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS audit_events_immutable_update
BEFORE UPDATE ON audit_events BEGIN
  SELECT RAISE(ABORT, 'audit events are immutable');
END;
CREATE TRIGGER IF NOT EXISTS audit_events_immutable_delete
BEFORE DELETE ON audit_events BEGIN
  SELECT RAISE(ABORT, 'audit events are immutable');
END;

CREATE TABLE IF NOT EXISTS transaction_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  category TEXT NOT NULL,
  item_name TEXT NOT NULL DEFAULT '',
  amount_cents INTEGER NOT NULL CHECK(amount_cents > 0),
  barber_id INTEGER REFERENCES users(id),
  item_type TEXT CHECK(item_type IN ('SERVICE', 'PRODUCT') OR item_type IS NULL) DEFAULT 'SERVICE',
  discount_amount INTEGER NOT NULL DEFAULT 0,
  commission_earned INTEGER NOT NULL DEFAULT 0,
  bundle_id INTEGER REFERENCES discounts_and_bundles(id)
);
CREATE INDEX IF NOT EXISTS idx_transaction_items_tx ON transaction_items(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_barber ON transaction_items(barber_id);

CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_update
BEFORE UPDATE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;
CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_delete
BEFORE DELETE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;

-- Backoffice tables
CREATE TABLE IF NOT EXISTS branches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  address TEXT NOT NULL DEFAULT '',
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE IF NOT EXISTS employee_profit_sharing_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  rule_id INTEGER NOT NULL REFERENCES branch_profit_sharing_rules(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users(id),
  percentage REAL NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_employee_profit_sharing_rule ON employee_profit_sharing_rules(rule_id);

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

CREATE TABLE IF NOT EXISTS password_resets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  used_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token_hash);
