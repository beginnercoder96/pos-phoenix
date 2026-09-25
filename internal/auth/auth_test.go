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
		username TEXT UNIQUE,
		email TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		active INTEGER NOT NULL DEFAULT 1,
		branch_id INTEGER,
		staff_type TEXT,
		phone_number TEXT,
		bank_name TEXT,
		bank_account_number TEXT,
		totp_secret TEXT,
		totp_enabled INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE sessions (
		token_hash TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE pre_auth_tokens (
		token_hash TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		failed_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until DATETIME,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		used_at DATETIME,
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

	// Seed admin with username 'admin' and a different DB password hash
	adminHash, err := HashPassword("some-other-secure-password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(username, email, display_name, password_hash, role, active) VALUES ('admin', 'admin@example.com', 'Administrator', ?, 'superadmin', 1)`, adminHash)
	if err != nil {
		t.Fatal(err)
	}

	// Seed ipang with username 'ipang' and another DB password hash
	ipangHash, err := HashPassword("ipang-original-password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(username, email, display_name, password_hash, role, active) VALUES ('ipang', 'ipang@example.com', 'Ipang', ?, 'superadmin', 1)`, ipangHash)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Username 'admin' with fixed password 'adminsupervisor'
	u1, err := svc.Authenticate(ctx, "admin", "adminsupervisor")
	if err != nil {
		t.Fatalf("admin username login failed: %v", err)
	}
	if u1.Role != "superadmin" || u1.Username != "admin" {
		t.Fatalf("admin user mismatch: %+v", u1)
	}

	// 2. Username 'ipang' with fixed password 'adminsupervisor'
	u2, err := svc.Authenticate(ctx, "ipang", "adminsupervisor")
	if err != nil {
		t.Fatalf("ipang username login failed: %v", err)
	}
	if u2.Role != "superadmin" || u2.Username != "ipang" {
		t.Fatalf("ipang user mismatch: %+v", u2)
	}

	// 3. Username 'yogi' with fixed password 'yogioperator' (auto-created on first login)
	u3, err := svc.Authenticate(ctx, "yogi", "yogioperator")
	if err != nil {
		t.Fatalf("yogi username login failed: %v", err)
	}
	if u3.Role != "operator" || u3.Username != "yogi" || u3.ID == 0 {
		t.Fatalf("yogi user mismatch: %+v", u3)
	}

	// Verify 'yogi' can be retrieved again
	u3Again, err := svc.Authenticate(ctx, "yogi", "yogioperator")
	if err != nil {
		t.Fatalf("yogi re-login failed: %v", err)
	}
	if u3Again.ID != u3.ID {
		t.Fatalf("expected same ID %d, got %d", u3.ID, u3Again.ID)
	}

	// 4. Case-insensitivity and trimming: '  ADMIN  '
	u1Upper, err := svc.Authenticate(ctx, "  ADMIN  ", "adminsupervisor")
	if err != nil {
		t.Fatalf("admin case-insensitive login failed: %v", err)
	}
	if u1Upper.ID != u1.ID {
		t.Fatalf("expected ID %d, got %d", u1.ID, u1Upper.ID)
	}

	// 5. Wrong passwords should fail
	if _, err := svc.Authenticate(ctx, "admin", "wrongpassword"); err == nil {
		t.Fatal("expected wrong password to fail for admin")
	}
	if _, err := svc.Authenticate(ctx, "yogi", "wrongpassword"); err == nil {
		t.Fatal("expected wrong password to fail for yogi")
	}

	// 6. Original DB password works with username 'admin'
	u1Orig, err := svc.Authenticate(ctx, "admin", "some-other-secure-password")
	if err != nil {
		t.Fatalf("admin original password failed: %v", err)
	}
	if u1Orig.ID != u1.ID {
		t.Fatalf("expected ID %d, got %d", u1.ID, u1Orig.ID)
	}

	// 7. Email login should succeed as well (both username and email are supported)
	u1Email, err := svc.Authenticate(ctx, "admin@example.com", "adminsupervisor")
	if err != nil {
		t.Fatalf("expected email login to succeed for admin@example.com, got: %v", err)
	}
	if u1Email.ID != u1.ID {
		t.Fatalf("expected ID %d, got %d", u1.ID, u1Email.ID)
	}

	u2Email, err := svc.Authenticate(ctx, "ipang@example.com", "adminsupervisor")
	if err != nil {
		t.Fatalf("expected email login to succeed for ipang@example.com, got: %v", err)
	}
	if u2Email.ID != u2.ID {
		t.Fatalf("expected ID %d, got %d", u2.ID, u2Email.ID)
	}

	u3Email, err := svc.Authenticate(ctx, "yogi@contoh.com", "yogioperator")
	if err != nil {
		t.Fatalf("expected email login to succeed for yogi@contoh.com, got: %v", err)
	}
	if u3Email.ID != u3.ID {
		t.Fatalf("expected ID %d, got %d", u3.ID, u3Email.ID)
	}

	// 8. Nonexistent user or wrong credentials should fail
	if _, err := svc.Authenticate(ctx, "nonexistent@example.com", "adminsupervisor"); err == nil {
		t.Fatal("expected nonexistent user to fail")
	}
	if _, err := svc.Authenticate(ctx, "", "adminsupervisor"); err == nil {
		t.Fatal("expected empty username to fail")
	}
}

func TestPasswordResetFlow(t *testing.T) {
	db := setupTestDB(t)
	svc := Service{DB: db}
	ctx := context.Background()

	// Seed user with email and username
	hash, err := HashPassword("initialPassword123!")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(username, email, display_name, password_hash, role, active) VALUES('testuser', 'testuser@example.com', 'Test User', ?, 'operator', 1)`, hash)
	if err != nil {
		t.Fatal(err)
	}

	// Request password reset
	token, resetURL, err := svc.CreatePasswordResetToken(ctx, "testuser@example.com")
	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}
	if token == "" || resetURL == "" {
		t.Fatalf("expected non-empty token and resetURL, got token=%s, url=%s", token, resetURL)
	}

	// Validate token
	u, err := svc.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidatePasswordResetToken failed: %v", err)
	}
	if u.Username != "testuser" || u.Email != "testuser@example.com" {
		t.Fatalf("user mismatch from ValidatePasswordResetToken: %+v", u)
	}

	// Reset password
	newPass := "NewSuperSecurePass456!"
	if err := svc.ResetPasswordWithToken(ctx, token, newPass); err != nil {
		t.Fatalf("ResetPasswordWithToken failed: %v", err)
	}

	// Login with new password should succeed using username
	uNew, err := svc.Authenticate(ctx, "testuser", newPass)
	if err != nil {
		t.Fatalf("Authenticate with new password failed: %v", err)
	}
	if uNew.Username != "testuser" {
		t.Fatalf("unexpected user: %+v", uNew)
	}

	// Old password must fail
	if _, err := svc.Authenticate(ctx, "testuser", "initialPassword123!"); err == nil {
		t.Fatal("expected old password to fail, but it succeeded")
	}

	// Token cannot be reused
	if err := svc.ResetPasswordWithToken(ctx, token, "YetAnotherPass789!"); err == nil {
		t.Fatal("expected reused token to fail, but it succeeded")
	}

	// Requesting for non-existent email should fail
	if _, _, err := svc.CreatePasswordResetToken(ctx, "nonexistent@example.com"); err == nil {
		t.Fatal("expected non-existent email to fail")
	}
}

func TestBootstrapAdminRequiresEmail(t *testing.T) {
	db := setupTestDB(t)
	svc := Service{DB: db}
	ctx := context.Background()

	// Initial system bootstrap requires admin email and password
	adminEmail := "superowner@example.com"
	adminPass := "superSecret123!"
	if err := svc.BootstrapAdmin(ctx, adminEmail, adminPass); err != nil {
		t.Fatalf("BootstrapAdmin failed: %v", err)
	}

	// Verify user is recorded in database with role 'superadmin' and extracted username
	var count int
	var uname, email, role string
	err := db.QueryRow(`SELECT COUNT(*), username, email, role FROM users WHERE email=?`, adminEmail).Scan(&count, &uname, &email, &role)
	if err != nil || count != 1 {
		t.Fatalf("expected admin created in DB: count=%d, err=%v", count, err)
	}
	if role != "superadmin" || email != adminEmail || uname != "superowner" {
		t.Fatalf("unexpected admin user data: role=%s, email=%s, uname=%s", role, email, uname)
	}

	// Daily routine login succeeds using username (not requiring email input)
	u, err := svc.Authenticate(ctx, "superowner", adminPass)
	if err != nil {
		t.Fatalf("expected routine login with username 'superowner' to succeed: %v", err)
	}
	if u.Role != "superadmin" {
		t.Fatalf("expected superadmin role, got %s", u.Role)
	}

	// Bootstrap is idempotent (running again does not recreate or fail)
	if err := svc.BootstrapAdmin(ctx, "another@example.com", "pass"); err != nil {
		t.Fatalf("expected subsequent BootstrapAdmin to safely no-op: %v", err)
	}
}

