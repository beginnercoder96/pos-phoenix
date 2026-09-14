package auth

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('superadmin','operator')),
		active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE sessions (
		token_hash TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestTemporaryFixedCredentials(t *testing.T) {
	db := setupTestDB(t)
	svc := Service{DB: db}
	ctx := context.Background()

	// Seed admin@example.com with a different DB password hash
	adminHash, err := HashPassword("some-other-secure-password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(email, display_name, password_hash, role, active) VALUES ('admin@example.com', 'Administrator', ?, 'superadmin', 1)`, adminHash)
	if err != nil {
		t.Fatal(err)
	}

	// Seed ipang@example.com with another DB password hash
	ipangHash, err := HashPassword("ipang-original-password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(email, display_name, password_hash, role, active) VALUES ('ipang@example.com', 'Ipang', ?, 'superadmin', 1)`, ipangHash)
	if err != nil {
		t.Fatal(err)
	}

	// 1. admin@example.com with adminsupervisor
	u1, err := svc.Authenticate(ctx, "admin@example.com", "adminsupervisor")
	if err != nil {
		t.Fatalf("admin@example.com login failed: %v", err)
	}
	if u1.Role != "superadmin" || u1.Email != "admin@example.com" {
		t.Fatalf("admin user mismatch: %+v", u1)
	}

	// 2. ipang@example.com with adminsupervisor
	u2, err := svc.Authenticate(ctx, "ipang@example.com", "adminsupervisor")
	if err != nil {
		t.Fatalf("ipang@example.com login failed: %v", err)
	}
	if u2.Role != "superadmin" || u2.Email != "ipang@example.com" {
		t.Fatalf("ipang user mismatch: %+v", u2)
	}

	// 3. yogi@contoh.com with yogioperator (not in DB yet)
	u3, err := svc.Authenticate(ctx, "yogi@contoh.com", "yogioperator")
	if err != nil {
		t.Fatalf("yogi@contoh.com login failed: %v", err)
	}
	if u3.Role != "operator" || u3.Email != "yogi@contoh.com" || u3.ID == 0 {
		t.Fatalf("yogi user mismatch: %+v", u3)
	}

	// Verify yogi@contoh.com can be retrieved again
	u3Again, err := svc.Authenticate(ctx, "yogi@contoh.com", "yogioperator")
	if err != nil {
		t.Fatalf("yogi@contoh.com re-login failed: %v", err)
	}
	if u3Again.ID != u3.ID {
		t.Fatalf("expected same ID %d, got %d", u3.ID, u3Again.ID)
	}

	// 4. Case-insensitivity and trimming
	u1Upper, err := svc.Authenticate(ctx, "  ADMIN@EXAMPLE.COM  ", "adminsupervisor")
	if err != nil {
		t.Fatalf("admin case-insensitive login failed: %v", err)
	}
	if u1Upper.ID != u1.ID {
		t.Fatalf("expected ID %d, got %d", u1.ID, u1Upper.ID)
	}

	// 5. Wrong passwords should fail
	if _, err := svc.Authenticate(ctx, "admin@example.com", "wrongpassword"); err == nil {
		t.Fatal("expected wrong password to fail for admin")
	}
	if _, err := svc.Authenticate(ctx, "yogi@contoh.com", "wrongpassword"); err == nil {
		t.Fatal("expected wrong password to fail for yogi")
	}

	// 6. Original DB password still works for admin
	u1Orig, err := svc.Authenticate(ctx, "admin@example.com", "some-other-secure-password")
	if err != nil {
		t.Fatalf("admin original password failed: %v", err)
	}
	if u1Orig.ID != u1.ID {
		t.Fatalf("expected ID %d, got %d", u1.ID, u1Orig.ID)
	}
}

