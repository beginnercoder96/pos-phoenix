package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID                       int64
	Email, DisplayName, Role string
	Active                   bool
}
type Service struct {
	DB  *sql.DB
	Now func() time.Time
}

func (s Service) ListOperators(ctx context.Context) ([]User, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,email,display_name,role,active FROM users WHERE role='operator' ORDER BY display_name,email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Active); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s Service) CreateOperator(ctx context.Context, email, displayName, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	if email == "" || !strings.Contains(email, "@") || len(email) > 254 {
		return errors.New("enter a valid email")
	}
	if displayName == "" || len(displayName) > 80 {
		return errors.New("display name is required")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO users(email,display_name,password_hash,role) VALUES(?,?,?,'operator')`, email, displayName, hash)
	if err != nil {
		return errors.New("an account with that email already exists")
	}
	return nil
}

func (s Service) SetOperatorActive(ctx context.Context, operatorID int64, active bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET active=? WHERE id=? AND role='operator'`, active, operatorID)
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
	Password    string
	DisplayName string
	Role        string
}

var temporaryFixedCredentials = map[string]fixedCredential{
	"admin@example.com": {Password: "adminsupervisor", DisplayName: "Administrator", Role: "superadmin"},
	"ipang@example.com": {Password: "adminsupervisor", DisplayName: "Ipang", Role: "superadmin"},
	"yogi@contoh.com":   {Password: "yogioperator", DisplayName: "yogi", Role: "operator"},
}

func (s Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	// Check temporary fixed credentials
	if fixed, ok := temporaryFixedCredentials[cleanEmail]; ok && fixed.Password == password {
		var u User
		err := s.DB.QueryRowContext(ctx, `SELECT id,email,display_name,role FROM users WHERE email=? AND active=1`, cleanEmail).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role)
		if err == nil {
			return u, nil
		}
		// If the user does not exist in DB yet, safely create it so foreign keys (sessions, transactions) work
		hash, _ := HashPassword(fixed.Password)
		res, err := s.DB.ExecContext(ctx, `INSERT INTO users(email,display_name,password_hash,role,active) VALUES(?,?,?,?,1)`, cleanEmail, fixed.DisplayName, hash, fixed.Role)
		if err == nil {
			id, _ := res.LastInsertId()
			return User{
				ID:          id,
				Email:       cleanEmail,
				DisplayName: fixed.DisplayName,
				Role:        fixed.Role,
				Active:      true,
			}, nil
		}
	}

	var u User
	var hash string
	err := s.DB.QueryRowContext(ctx, `SELECT id,email,display_name,role,password_hash FROM users WHERE email=? AND active=1`, cleanEmail).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &hash)
	if err != nil || !VerifyPassword(hash, password) {
		return User{}, errors.New("invalid credentials")
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
	err := s.DB.QueryRowContext(ctx, `SELECT u.id,u.email,u.display_name,u.role FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND u.active=1`, base64.RawStdEncoding.EncodeToString(sum[:]), s.now().UTC()).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role)
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
	_, err = s.DB.ExecContext(ctx, `INSERT INTO users(email,display_name,password_hash,role) VALUES(?,?,?,'superadmin')`, strings.ToLower(strings.TrimSpace(email)), "Administrator", hash)
	return err
}
func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
