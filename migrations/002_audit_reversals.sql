PRAGMA foreign_keys = ON;

ALTER TABLE transactions ADD COLUMN reversal_of_id INTEGER REFERENCES transactions(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_one_reversal
  ON transactions(reversal_of_id) WHERE reversal_of_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_id INTEGER NOT NULL REFERENCES users(id),
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  details TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_events_created
  ON audit_events(created_at DESC, id DESC);

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
