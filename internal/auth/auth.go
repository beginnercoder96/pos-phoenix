package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID                       int64
	Username                 string
	Email, DisplayName, Role string
	Active                   bool
	BranchID                 int64
	StaffType                string
	PhoneNumber              string
	BankName                 string
	BankAccountNumber        string
}

type Service struct {
	DB  *sql.DB
	Now func() time.Time
}

func (s Service) ListOperators(ctx context.Context) ([]User, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,COALESCE(username,''),email,display_name,role,active,COALESCE(branch_id,0),COALESCE(staff_type,''),COALESCE(phone_number,''),COALESCE(bank_name,''),COALESCE(bank_account_number,'') FROM users WHERE role IN ('operator','barberman','cashier') ORDER BY display_name,email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.Active, &user.BranchID, &user.StaffType, &user.PhoneNumber, &user.BankName, &user.BankAccountNumber); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// ListEmployees returns active employees (barbermen/cashiers) for a specific branch.
func (s Service) ListEmployees(ctx context.Context, branchID int64) ([]User, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id,COALESCE(username,''),email,display_name,role,active,COALESCE(branch_id,0),COALESCE(staff_type,''),COALESCE(phone_number,''),COALESCE(bank_name,''),COALESCE(bank_account_number,'')
		 FROM users WHERE branch_id=? AND active=1
		 AND role IN ('operator','barberman','cashier')
		 ORDER BY display_name,email`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.Active, &user.BranchID, &user.StaffType, &user.PhoneNumber, &user.BankName, &user.BankAccountNumber); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s Service) CreateOperator(ctx context.Context, email, displayName, password string) error {
	uname := email
	if at := strings.Index(email, "@"); at > 0 {
		uname = email[:at]
	}
	return s.CreateOperatorWithBranchAndCredentials(ctx, uname, email, displayName, password, "operator", "barberman", 0, "", "", "")
}

func (s Service) CreateOperatorWithBranch(ctx context.Context, email, displayName, password, role, staffType string, branchID int64) error {
	uname := email
	if at := strings.Index(email, "@"); at > 0 {
		uname = email[:at]
	}
	return s.CreateOperatorWithBranchAndCredentials(ctx, uname, email, displayName, password, role, staffType, branchID, "", "", "")
}

func (s Service) CreateOperatorWithBranchAndCredentials(ctx context.Context, username, email, displayName, password, role, staffType string, branchID int64, phoneNumber, bankName, bankAccountNumber string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	phoneNumber = strings.TrimSpace(phoneNumber)
	bankName = strings.TrimSpace(bankName)
	bankAccountNumber = strings.TrimSpace(bankAccountNumber)

	if username == "" {
		if at := strings.Index(email, "@"); at > 0 {
			username = strings.ToLower(strings.TrimSpace(email[:at]))
		} else {
			username = strings.ToLower(strings.TrimSpace(displayName))
		}
	}
	username = strings.ReplaceAll(username, " ", "_")
	if len(username) < 3 || len(username) > 50 {
		return errors.New("username must be 3 to 50 characters")
	}
	if email == "" || !strings.Contains(email, "@") || len(email) > 254 {
		return errors.New("enter a valid email")
	}
	if displayName == "" || len(displayName) > 80 {
		return errors.New("display name is required")
	}
	if role == "" {
		role = "operator"
	}
	if staffType == "" {
		staffType = "barberman"
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	var bID any = branchID
	if branchID <= 0 {
		bID = nil
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO users(username,email,display_name,password_hash,role,branch_id,staff_type,phone_number,bank_name,bank_account_number) VALUES(?,?,?,?,?,?,?,?,?,?)`, username, email, displayName, hash, role, bID, staffType, phoneNumber, bankName, bankAccountNumber)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "idx_users_username") || strings.Contains(errStr, "users.username") {
			return errors.New("an account with that username already exists")
		}
		return errors.New("an account with that email already exists")
	}
	return nil
}

func (s Service) UpdateOperatorCredentials(ctx context.Context, operatorID int64, displayName, phoneNumber, bankName, bankAccountNumber string, branchID int64, staffType string) error {
	displayName = strings.TrimSpace(displayName)
	phoneNumber = strings.TrimSpace(phoneNumber)
	bankName = strings.TrimSpace(bankName)
	bankAccountNumber = strings.TrimSpace(bankAccountNumber)
	if displayName == "" || len(displayName) > 80 {
		return errors.New("display name is required")
	}
	var bID any = branchID
	if branchID <= 0 {
		bID = nil
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE users SET display_name=?, phone_number=?, bank_name=?, bank_account_number=?, branch_id=?, staff_type=? WHERE id=? AND role IN ('operator','barberman','cashier')`, displayName, phoneNumber, bankName, bankAccountNumber, bID, staffType, operatorID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return errors.New("operator not found")
	}
	return nil
}

func (s Service) SetOperatorActive(ctx context.Context, operatorID int64, active bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET active=? WHERE id=? AND role IN ('operator','barberman','cashier')`, active, operatorID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return errors.New("operator not found")
	}
	if !active {
		if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, operatorID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type fixedCredential struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
	Role        string
	StaffType   string
}

var temporaryFixedCredentials = map[string]fixedCredential{
	"admin":             {Username: "admin", Email: "admin@example.com", Password: "adminsupervisor", DisplayName: "Administrator", Role: "superadmin", StaffType: "owner"},
	"admin@example.com": {Username: "admin", Email: "admin@example.com", Password: "adminsupervisor", DisplayName: "Administrator", Role: "superadmin", StaffType: "owner"},
	"ipang":             {Username: "ipang", Email: "ipang@example.com", Password: "adminsupervisor", DisplayName: "Ipang", Role: "superadmin", StaffType: "owner"},
	"ipang@example.com": {Username: "ipang", Email: "ipang@example.com", Password: "adminsupervisor", DisplayName: "Ipang", Role: "superadmin", StaffType: "owner"},
	"yogi":              {Username: "yogi", Email: "yogi@contoh.com", Password: "yogioperator", DisplayName: "yogi", Role: "operator", StaffType: "barberman"},
	"yogi@contoh.com":   {Username: "yogi", Email: "yogi@contoh.com", Password: "yogioperator", DisplayName: "yogi", Role: "operator", StaffType: "barberman"},
}

func (s Service) Authenticate(ctx context.Context, username, password string) (User, error) {
	cleanID := strings.ToLower(strings.TrimSpace(username))
	if cleanID == "" {
		return User{}, errors.New("invalid credentials")
	}

	// Check temporary fixed credentials (supports either username or email)
	if fixed, ok := temporaryFixedCredentials[cleanID]; ok && fixed.Password == password {
		var u User
		err := s.DB.QueryRowContext(ctx, `SELECT id,COALESCE(username,''),email,display_name,role,COALESCE(branch_id,0),COALESCE(staff_type,'') FROM users WHERE (LOWER(username)=? OR LOWER(email)=? OR LOWER(username)=? OR LOWER(email)=?) AND active=1`, cleanID, cleanID, fixed.Username, fixed.Email).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &u.BranchID, &u.StaffType)
		if err == nil {
			if u.Username == "" {
				u.Username = fixed.Username
			}
			return u, nil
		}
		// If the user does not exist in DB yet, safely create it so foreign keys (sessions, transactions) work
		hash, _ := HashPassword(fixed.Password)
		res, err := s.DB.ExecContext(ctx, `INSERT INTO users(username,email,display_name,password_hash,role,active,staff_type) VALUES(?,?,?,?,?,1,?)`, fixed.Username, fixed.Email, fixed.DisplayName, hash, fixed.Role, fixed.StaffType)
		if err == nil {
			id, _ := res.LastInsertId()
			return User{
				ID:          id,
				Username:    fixed.Username,
				Email:       fixed.Email,
				DisplayName: fixed.DisplayName,
				Role:        fixed.Role,
				Active:      true,
				StaffType:   fixed.StaffType,
			}, nil
		}
	}

	var u User
	var hash string
	err := s.DB.QueryRowContext(ctx, `SELECT id,COALESCE(username,''),email,display_name,role,password_hash,COALESCE(branch_id,0),COALESCE(staff_type,'') FROM users WHERE (LOWER(username)=? OR LOWER(email)=?) AND active=1`, cleanID, cleanID).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &hash, &u.BranchID, &u.StaffType)
	if err != nil || !VerifyPassword(hash, password) {
		return User{}, errors.New("invalid credentials")
	}
	if u.Username == "" {
		if strings.Contains(cleanID, "@") {
			u.Username = strings.Split(cleanID, "@")[0]
		} else {
			u.Username = cleanID
		}
	}
	return u, nil
}

func (s Service) CreateSession(ctx context.Context, userID int64) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	_, err := s.DB.ExecContext(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at) VALUES(?,?,?)`, base64.RawStdEncoding.EncodeToString(sum[:]), userID, s.now().Add(12*time.Hour).UTC())
	return token, err
}

func (s Service) UserForSession(ctx context.Context, token string) (User, error) {
	sum := sha256.Sum256([]byte(token))
	var u User
	err := s.DB.QueryRowContext(ctx, `SELECT u.id,COALESCE(u.username,''),u.email,u.display_name,u.role,COALESCE(u.branch_id,0),COALESCE(u.staff_type,'') FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND u.active=1`, base64.RawStdEncoding.EncodeToString(sum[:]), s.now().UTC()).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &u.BranchID, &u.StaffType)
	return u, err
}

func (s Service) DeleteSession(ctx context.Context, token string) error {
	sum := sha256.Sum256([]byte(token))
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash=?`, base64.RawStdEncoding.EncodeToString(sum[:]))
	return err
}

func (s Service) BootstrapAdmin(ctx context.Context, email, password string) error {
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role='superadmin'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	uname := "admin"
	if at := strings.Index(email, "@"); at > 0 {
		uname = email[:at]
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO users(username,email,display_name,password_hash,role) VALUES(?,?,?,?,'superadmin')`, uname, strings.ToLower(strings.TrimSpace(email)), "Administrator", hash)
	return err
}

// CreatePasswordResetToken generates a secure token and saves it to password_resets table.
// It also logs the mock email output to stdout for development.
func (s Service) CreatePasswordResetToken(ctx context.Context, email string) (token string, resetURL string, err error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var u User
	err = s.DB.QueryRowContext(ctx, `SELECT id, COALESCE(username,''), email, display_name, role FROM users WHERE LOWER(email)=? AND active=1`, cleanEmail).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role)
	if err != nil {
		return "", "", errors.New("no active account found with that email")
	}

	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	tokenHash := base64.RawStdEncoding.EncodeToString(sum[:])
	expiresAt := s.now().Add(1 * time.Hour).UTC()

	_, err = s.DB.ExecContext(ctx, `INSERT INTO password_resets(user_id, token_hash, expires_at) VALUES(?,?,?)`, u.ID, tokenHash, expiresAt)
	if err != nil {
		return "", "", err
	}

	resetURL = "/reset-password?token=" + token

	// Mock email delivery printed clearly to console/log
	fmt.Printf("\n========================================\n[MOCK EMAIL SERVICE]\nTo: %s\nSubject: Atur Ulang Kata Sandi POS Phoenix\nHalo %s (@%s),\nKlik tautan berikut untuk mengatur ulang kata sandi Anda:\n%s\n(Tautan berlaku selama 1 jam)\n========================================\n\n", u.Email, u.DisplayName, u.Username, resetURL)

	return token, resetURL, nil
}

// ValidatePasswordResetToken validates that the token exists, is not expired, and has not been used.
func (s Service) ValidatePasswordResetToken(ctx context.Context, token string) (User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, errors.New("invalid or missing reset token")
	}
	sum := sha256.Sum256([]byte(token))
	tokenHash := base64.RawStdEncoding.EncodeToString(sum[:])

	var u User
	var expiresAt time.Time
	var usedAt sql.NullTime
	err := s.DB.QueryRowContext(ctx, `
		SELECT u.id, COALESCE(u.username,''), u.email, u.display_name, u.role, pr.expires_at, pr.used_at
		FROM password_resets pr
		JOIN users u ON u.id = pr.user_id
		WHERE pr.token_hash = ? AND u.active = 1
	`, tokenHash).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &expiresAt, &usedAt)
	if err != nil {
		return User{}, errors.New("invalid or expired reset link")
	}
	if usedAt.Valid {
		return User{}, errors.New("this reset link has already been used")
	}
	if s.now().UTC().After(expiresAt) {
		return User{}, errors.New("this reset link has expired")
	}
	return u, nil
}

// ResetPasswordWithToken changes the user's password and marks the token used.
func (s Service) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	u, err := s.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	tokenHash := base64.RawStdEncoding.EncodeToString(sum[:])

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update password
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, u.ID); err != nil {
		return err
	}

	// Mark token used
	if _, err := tx.ExecContext(ctx, `UPDATE password_resets SET used_at = ? WHERE token_hash = ?`, s.now().UTC(), tokenHash); err != nil {
		return err
	}

	// Invalidate previous sessions so user must log in with new password
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, u.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

