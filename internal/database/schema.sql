PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT,email TEXT NOT NULL UNIQUE,display_name TEXT NOT NULL,password_hash TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN ('superadmin','operator')),active INTEGER NOT NULL DEFAULT 1,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS sessions (token_hash TEXT PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,expires_at DATETIME NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expires_at);
CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY AUTOINCREMENT,operator_id INTEGER NOT NULL REFERENCES users(id),kind TEXT NOT NULL CHECK(kind IN ('income','expense')),amount_cents INTEGER NOT NULL CHECK(amount_cents > 0),category TEXT NOT NULL,note TEXT NOT NULL DEFAULT '',occurred_at DATETIME NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,reversal_of_id INTEGER REFERENCES transactions(id));
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_transactions_operator_date ON transactions(operator_id,occurred_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_one_reversal ON transactions(reversal_of_id) WHERE reversal_of_id IS NOT NULL;

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
  amount_cents INTEGER NOT NULL CHECK(amount_cents > 0)
);
CREATE INDEX IF NOT EXISTS idx_transaction_items_tx ON transaction_items(transaction_id);

CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_update
BEFORE UPDATE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;
CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_delete
BEFORE DELETE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;
