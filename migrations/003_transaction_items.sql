PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS transaction_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  category TEXT NOT NULL,
  item_name TEXT NOT NULL DEFAULT '',
  amount_cents INTEGER NOT NULL CHECK(amount_cents > 0)
);

CREATE INDEX IF NOT EXISTS idx_transaction_items_tx
  ON transaction_items(transaction_id);

CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_update
BEFORE UPDATE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;

CREATE TRIGGER IF NOT EXISTS transaction_items_immutable_delete
BEFORE DELETE ON transaction_items BEGIN
  SELECT RAISE(ABORT, 'transaction items are immutable');
END;

