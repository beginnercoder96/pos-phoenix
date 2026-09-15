package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err = ensureTransactionReversalColumn(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err = ensureBackofficeColumns(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate backoffice columns: %w", err)
	}
	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; " + schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	if err = seedBranches(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("seed branches: %w", err)
	}
	return db, nil
}

func ensureTransactionReversalColumn(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operator_id INTEGER NOT NULL REFERENCES users(id),
		kind TEXT NOT NULL CHECK (kind IN ('income','expense')),
		amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
		category TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		occurred_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		reversal_of_id INTEGER REFERENCES transactions(id)
	)`); err != nil {
		return err
	}
	rows, err := db.Query(`PRAGMA table_info(transactions)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	hasColumn := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == "reversal_of_id" {
			hasColumn = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !hasColumn {
		_, err = db.Exec(`ALTER TABLE transactions ADD COLUMN reversal_of_id INTEGER REFERENCES transactions(id)`)
	}
	return err
}

// ensureBackofficeColumns adds new columns needed for multi-branch and barber tracking.
func ensureBackofficeColumns(db *sql.DB) error {
	// Add branch_id to transactions if missing
	if !hasColumn(db, "transactions", "branch_id") {
		if _, err := db.Exec(`ALTER TABLE transactions ADD COLUMN branch_id INTEGER REFERENCES branches(id)`); err != nil {
			// Ignore error if column already exists (race condition with schema.sql)
			_ = err
		}
	}

	// Add branch_id and staff_type to users if missing
	if !hasColumn(db, "users", "branch_id") {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN branch_id INTEGER REFERENCES branches(id)`); err != nil {
			_ = err
		}
	}
	if !hasColumn(db, "users", "staff_type") {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN staff_type TEXT`); err != nil {
			_ = err
		}
	}

	// Add new columns to transaction_items if missing
	for _, col := range []struct {
		name string
		def  string
	}{
		{"barber_id", "INTEGER REFERENCES users(id)"},
		{"item_type", "TEXT DEFAULT 'SERVICE'"},
		{"discount_amount", "INTEGER NOT NULL DEFAULT 0"},
		{"commission_earned", "INTEGER NOT NULL DEFAULT 0"},
		{"bundle_id", "INTEGER REFERENCES discounts_and_bundles(id)"},
	} {
		if !hasColumn(db, "transaction_items", col.name) {
			if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE transaction_items ADD COLUMN %s %s`, col.name, col.def)); err != nil {
				_ = err
			}
		}
	}

	return nil
}

// hasColumn checks if a column exists in a table.
func hasColumn(db *sql.DB, table, column string) bool {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

// seedBranches inserts the two Pardis Barbershop branches and initial backoffice data if missing.
func seedBranches(db *sql.DB) error {
	branches := []struct {
		code, name, address string
	}{
		{"KLASEMAN", "Pardis Barbershop Klaseman", "Klaseman"},
		{"LEDOK", "Pardis Barbershop Ledok", "Ledok"},
	}
	for _, b := range branches {
		_, err := db.Exec(
			`INSERT OR IGNORE INTO branches (code, name, address) VALUES (?, ?, ?)`,
			b.code, b.name, b.address)
		if err != nil {
			return err
		}
	}

	// Seed catalog items if empty
	var catCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM catalog_items`).Scan(&catCount); err == nil && catCount == 0 {
		items := []struct {
			name, category, itemType string
			price, commission       int64
		}{
			{"Haircut Regular", "Haircut Services", "SERVICE", 4000000, 0},
			{"Haircut + Wash", "Haircut Services", "SERVICE", 6000000, 0},
			{"Beard Trim & Shave", "Grooming", "SERVICE", 2500000, 0},
			{"Hair Treatment & Color", "Treatment", "SERVICE", 10000000, 0},
			{"Pomade Death Waterbased", "Product", "PRODUCT", 8500000, 500000},
			{"Hair Tonic Ginseng", "Product", "PRODUCT", 6500000, 1000000},
			{"Hair Clay Matte Finish", "Product", "PRODUCT", 7500000, 750000},
		}
		for _, it := range items {
			_, _ = db.Exec(`INSERT INTO catalog_items(name, category, item_type, price_cents, commission_amount, is_active) VALUES (?, ?, ?, ?, ?, 1)`,
				it.name, it.category, it.itemType, it.price, it.commission)
		}
	}

	// Seed discounts if empty
	var discCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM discounts_and_bundles`).Scan(&discCount); err == nil && discCount == 0 {
		_, _ = db.Exec(`INSERT INTO discounts_and_bundles(code, name, type, value, service_allocation_ratio, product_allocation_ratio, is_active) VALUES
			('PAKET-POMADE', 'Haircut + Pomade Bundling', 'BUNDLE', 11000000, 0.6, 0.4, 1),
			('PROMO-HEMAT', 'Diskon Hemat 10%', 'PERCENTAGE', 10, 1.0, 0.0, 1)`)
	}

	// Get branch IDs
	var klasemanID, ledokID int64
	_ = db.QueryRow(`SELECT id FROM branches WHERE code='KLASEMAN'`).Scan(&klasemanID)
	_ = db.QueryRow(`SELECT id FROM branches WHERE code='LEDOK'`).Scan(&ledokID)

	if klasemanID > 0 && ledokID > 0 {
		// Seed sample employees for Klaseman
		var empCountK int
		_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE branch_id=?`, klasemanID).Scan(&empCountK)
		if empCountK == 0 {
			_, _ = db.Exec(`INSERT OR IGNORE INTO users(email, display_name, password_hash, role, branch_id, staff_type, active) VALUES
				('karyawan1.klaseman@pardis.com', 'Karyawan 1 (Klaseman)', '$argon2id$v=19$m=65536,t=3,p=2$c2FtcGxlc2FsdDEyMzQ1Ng$c2FtcGxlaGFzaDEyMzQ1Njc4OTA=', 'operator', ?, 'barberman', 1),
				('karyawan2.klaseman@pardis.com', 'Karyawan 2 (Klaseman)', '$argon2id$v=19$m=65536,t=3,p=2$c2FtcGxlc2FsdDEyMzQ1Ng$c2FtcGxlaGFzaDEyMzQ1Njc4OTA=', 'operator', ?, 'barberman', 1)`,
				klasemanID, klasemanID)
		}

		// Seed sample employees for Ledok
		var empCountL int
		_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE branch_id=?`, ledokID).Scan(&empCountL)
		if empCountL == 0 {
			_, _ = db.Exec(`INSERT OR IGNORE INTO users(email, display_name, password_hash, role, branch_id, staff_type, active) VALUES
				('karyawan1.ledok@pardis.com', 'Karyawan 1 (Ledok)', '$argon2id$v=19$m=65536,t=3,p=2$c2FtcGxlc2FsdDEyMzQ1Ng$c2FtcGxlaGFzaDEyMzQ1Njc4OTA=', 'operator', ?, 'barberman', 1),
				('karyawan2.ledok@pardis.com', 'Karyawan 2 (Ledok)', '$argon2id$v=19$m=65536,t=3,p=2$c2FtcGxlc2FsdDEyMzQ1Ng$c2FtcGxlaGFzaDEyMzQ1Njc4OTA=', 'operator', ?, 'barberman', 1)`,
				ledokID, ledokID)
		}

		// Seed profit sharing rules for current month and 2026-01
		periods := []string{"2026-01", time.Now().Format("2006-01")}
		for _, period := range periods {
			seedProfitSharingForPeriod(db, klasemanID, ledokID, period)
		}

		// Seed simulation transactions for 2026-01 and recent months if branch transactions are empty
		seedSimulationTransactions(db, klasemanID, ledokID)
	}

	return nil
}

func seedProfitSharingForPeriod(db *sql.DB, klasemanID, ledokID int64, period string) {
	// Klaseman: Owner 20%, Unallocated 10%, Karyawan 1 35%, Karyawan 2 35%
	_, _ = db.Exec(`INSERT OR IGNORE INTO branch_profit_sharing_rules(branch_id, period_month, owner_percentage, unallocated_percentage) VALUES(?, ?, 20.0, 10.0)`, klasemanID, period)
	var kRuleID int64
	if err := db.QueryRow(`SELECT id FROM branch_profit_sharing_rules WHERE branch_id=? AND period_month=?`, klasemanID, period).Scan(&kRuleID); err == nil {
		rows, _ := db.Query(`SELECT id FROM users WHERE branch_id=? AND role IN ('operator','barberman') ORDER BY id LIMIT 2`, klasemanID)
		var kEmps []int64
		if rows != nil {
			for rows.Next() {
				var id int64
				_ = rows.Scan(&id)
				kEmps = append(kEmps, id)
			}
			rows.Close()
		}
		for _, eid := range kEmps {
			_, _ = db.Exec(`INSERT OR IGNORE INTO employee_profit_sharing_rules(rule_id, user_id, percentage) VALUES(?, ?, 35.0)`, kRuleID, eid)
		}
	}

	// Ledok: Owner 20%, Unallocated 10%, Karyawan 1 40%, Karyawan 2 30%
	_, _ = db.Exec(`INSERT OR IGNORE INTO branch_profit_sharing_rules(branch_id, period_month, owner_percentage, unallocated_percentage) VALUES(?, ?, 20.0, 10.0)`, ledokID, period)
	var lRuleID int64
	if err := db.QueryRow(`SELECT id FROM branch_profit_sharing_rules WHERE branch_id=? AND period_month=?`, ledokID, period).Scan(&lRuleID); err == nil {
		rows, _ := db.Query(`SELECT id FROM users WHERE branch_id=? AND role IN ('operator','barberman') ORDER BY id LIMIT 2`, ledokID)
		var lEmps []int64
		if rows != nil {
			for rows.Next() {
				var id int64
				_ = rows.Scan(&id)
				lEmps = append(lEmps, id)
			}
			rows.Close()
		}
		if len(lEmps) >= 1 {
			_, _ = db.Exec(`INSERT OR IGNORE INTO employee_profit_sharing_rules(rule_id, user_id, percentage) VALUES(?, ?, 40.0)`, lRuleID, lEmps[0])
		}
		if len(lEmps) >= 2 {
			_, _ = db.Exec(`INSERT OR IGNORE INTO employee_profit_sharing_rules(rule_id, user_id, percentage) VALUES(?, ?, 30.0)`, lRuleID, lEmps[1])
		}
	}
}

func seedSimulationTransactions(db *sql.DB, klasemanID, ledokID int64) {
	var txCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE branch_id IS NOT NULL`).Scan(&txCount)
	if txCount > 0 {
		return
	}

	// Fetch employee IDs
	var k1, k2, l1, l2 int64
	_ = db.QueryRow(`SELECT id FROM users WHERE email='karyawan1.klaseman@pardis.com'`).Scan(&k1)
	_ = db.QueryRow(`SELECT id FROM users WHERE email='karyawan2.klaseman@pardis.com'`).Scan(&k2)
	_ = db.QueryRow(`SELECT id FROM users WHERE email='karyawan1.ledok@pardis.com'`).Scan(&l1)
	_ = db.QueryRow(`SELECT id FROM users WHERE email='karyawan2.ledok@pardis.com'`).Scan(&l2)

	// Seed January 2026 transactions matching the exact specification:
	// Klaseman: Omzet Jasa Rp 10.000.000, 10 Pomade sold by Karyawan 1 (Rp 850.000, komisi Rp 50.000), 5 Tonic by Karyawan 2 (Rp 325.000, komisi Rp 50.000)
	if k1 > 0 && k2 > 0 {
		// Service income: Rp 10.000.000 = 1000000000 cents
		resK, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', 1000000000, 'Haircut Services', 'Total Service Jan 2026', '2026-01-20 12:00:00', ?)`, k1, klasemanID)
		if err == nil {
			tID, _ := resK.LastInsertId()
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type)
				VALUES(?, 'Haircut Services', 'Service Omzet', 1000000000, 'SERVICE')`, tID)
		}

		// Karyawan 1: 10 Pomade @ Rp 85.000 = Rp 850.000 (85000000 cents), commission Rp 50.000 (5000000 cents)
		resP1, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', 85000000, 'Product', '10 Pomade by Karyawan 1', '2026-01-22 14:00:00', ?)`, k1, klasemanID)
		if err == nil {
			tID, _ := resP1.LastInsertId()
			for i := 0; i < 10; i++ {
				_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type, barber_id, commission_earned)
					VALUES(?, 'Product', 'Pomade Death Waterbased', 8500000, 'PRODUCT', ?, 500000)`, tID, k1)
			}
		}

		// Karyawan 2: 5 Tonic @ Rp 65.000 = Rp 325.000 (32500000 cents), commission Rp 50.000 (5000000 cents)
		resP2, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', 32500000, 'Product', '5 Tonic by Karyawan 2', '2026-01-23 15:00:00', ?)`, k2, klasemanID)
		if err == nil {
			tID, _ := resP2.LastInsertId()
			for i := 0; i < 5; i++ {
				_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type, barber_id, commission_earned)
					VALUES(?, 'Product', 'Hair Tonic Ginseng', 6500000, 'PRODUCT', ?, 1000000)`, tID, k2)
			}
		}
	}

	// Ledok: Omzet Jasa Rp 12.000.000, 8 Hair Clay by Karyawan 1 (Rp 600.000, komisi Rp 60.000)
	if l1 > 0 && l2 > 0 {
		// Service income: Rp 12.000.000 = 1200000000 cents
		resL, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', 1200000000, 'Haircut Services', 'Total Service Jan 2026', '2026-01-20 12:00:00', ?)`, l1, ledokID)
		if err == nil {
			tID, _ := resL.LastInsertId()
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type)
				VALUES(?, 'Haircut Services', 'Service Omzet', 1200000000, 'SERVICE')`, tID)
		}

		// Karyawan 1: 8 Hair Clay @ Rp 75.000 = Rp 600.000 (60000000 cents), commission Rp 60.000 (6000000 cents)
		resPL, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', 60000000, 'Product', '8 Hair Clay by Karyawan 1', '2026-01-24 16:00:00', ?)`, l1, ledokID)
		if err == nil {
			tID, _ := resPL.LastInsertId()
			for i := 0; i < 8; i++ {
				_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type, barber_id, commission_earned)
					VALUES(?, 'Product', 'Hair Clay Matte Finish', 7500000, 'PRODUCT', ?, 750000)`, tID, l1)
			}
		}
	}

	// Seed historical 24-month trend transactions
	now := time.Now()
	for i := 23; i >= 0; i-- {
		m := now.AddDate(0, -i, 0)
		dateStr := m.Format("2006-01") + "-15 11:00:00"

		// Klaseman monthly service + product
		svcK := int64(800000000 + ((i * 17) % 350000000))
		prdK := int64(40000000 + ((i * 11) % 60000000))
		resK, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', ?, 'Haircut Services', 'Monthly Service', ?, ?)`, k1, svcK+prdK, dateStr, klasemanID)
		if err == nil {
			tID, _ := resK.LastInsertId()
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type)
				VALUES(?, 'Haircut Services', 'Haircut', ?, 'SERVICE')`, tID, svcK)
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type, barber_id, commission_earned)
				VALUES(?, 'Product', 'Pomade Death Waterbased', ?, 'PRODUCT', ?, 500000)`, tID, prdK, k1)
		}

		// Ledok monthly service + product
		svcL := int64(900000000 + ((i * 23) % 400000000))
		prdL := int64(50000000 + ((i * 13) % 70000000))
		resL, err := db.Exec(`INSERT INTO transactions(operator_id, kind, amount_cents, category, note, occurred_at, branch_id)
			VALUES(?, 'income', ?, 'Haircut Services', 'Monthly Service', ?, ?)`, l1, svcL+prdL, dateStr, ledokID)
		if err == nil {
			tID, _ := resL.LastInsertId()
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type)
				VALUES(?, 'Haircut Services', 'Haircut', ?, 'SERVICE')`, tID, svcL)
			_, _ = db.Exec(`INSERT INTO transaction_items(transaction_id, category, item_name, amount_cents, item_type, barber_id, commission_earned)
				VALUES(?, 'Product', 'Hair Clay Matte Finish', ?, 'PRODUCT', ?, 750000)`, tID, prdL, l1)
		}
	}
}
