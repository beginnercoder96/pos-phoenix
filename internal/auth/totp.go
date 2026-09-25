package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// DevMasterOTP is a convenience master code for local development and testing,
// allowing developers to bypass scanning Google Authenticator on their physical phones.
const DevMasterOTP = "123456"
const DevMasterOTPAlt = "000000"

// GenerateTOTPSecret generates a cryptographically secure 20-byte random secret,
// Base32 encoded without padding (RFC 6238 standard).
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateTOTPURI creates the standard otpauth URI compatible with Google Authenticator and Authy.
// Format: otpauth://totp/POS%20Phoenix:{username}?secret={secret}&issuer=POS%20Phoenix
func GenerateTOTPURI(username, secret, issuer string) string {
	if issuer == "" {
		issuer = "POS Phoenix"
	}
	label := fmt.Sprintf("%s:%s", issuer, username)
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", "30")

	return fmt.Sprintf("otpauth://totp/%s?%s", url.PathEscape(label), q.Encode())
}

// GenerateQRCodeDataURL generates a base64-encoded PNG Data URL from a TOTP URI.
func GenerateQRCodeDataURL(uri string) (string, error) {
	png, err := qrcode.Encode(uri, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("encode qr code: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// GenerateTOTPCode computes the 6-digit TOTP code for a given secret at a specific time (RFC 6238).
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	secretClean := strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretClean)
	if err != nil {
		key, err = base32.StdEncoding.DecodeString(secretClean)
		if err != nil {
			return "", fmt.Errorf("decode base32 secret: %w", err)
		}
	}

	counter := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binaryCode := binary.BigEndian.Uint32(sum[offset : offset+4])
	truncated := binaryCode & 0x7fffffff

	codeInt := truncated % 1000000
	return fmt.Sprintf("%06d", codeInt), nil
}

// ValidateTOTPCode validates the 6-digit code against the secret key.
// It checks t-30s, t, and t+30s to allow a ±1 time-step skew (RFC 6238 tolerance).
// If allowDevBypass is true (development mode), it also accepts DevMasterOTP ("123456" / "000000").
// In production (allowDevBypass=false), master codes are strictly rejected.
func ValidateTOTPCode(secret, code string, t time.Time, allowDevBypass bool) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}

	// Developer bypass for effortless local testing (disabled in production)
	if allowDevBypass && (code == DevMasterOTP || code == DevMasterOTPAlt) {
		return true
	}

	timeSteps := []time.Time{
		t.Add(-30 * time.Second),
		t,
		t.Add(30 * time.Second),
	}

	for _, stepTime := range timeSteps {
		generated, err := GenerateTOTPCode(secret, stepTime)
		if err == nil && generated == code {
			return true
		}
	}

	return false
}

// FormatSecretForDisplay groups the base32 secret in chunks of 4 for human readability.
func FormatSecretForDisplay(secret string) string {
	secret = strings.TrimSpace(secret)
	var parts []string
	for i := 0; i < len(secret); i += 4 {
		end := i + 4
		if end > len(secret) {
			end = len(secret)
		}
		parts = append(parts, secret[i:end])
	}
	return strings.Join(parts, " ")
}
