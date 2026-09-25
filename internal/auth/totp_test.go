package auth

import (
	"strings"
	"testing"
	"time"
)

func TestTOTPGenerationAndValidation(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}
	if len(secret) < 16 {
		t.Fatalf("secret too short: %s", secret)
	}

	now := time.Now()
	code, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("code must be 6 digits, got: %s", code)
	}

	// 1. Valid code at current time
	if !ValidateTOTPCode(secret, code, now, false) {
		t.Fatalf("expected valid code %s to validate successfully", code)
	}

	// 2. Tolerance for ±30s skew
	pastCode, _ := GenerateTOTPCode(secret, now.Add(-30*time.Second))
	if !ValidateTOTPCode(secret, pastCode, now, false) {
		t.Fatalf("expected past 30s code to be accepted within skew window")
	}

	futureCode, _ := GenerateTOTPCode(secret, now.Add(30*time.Second))
	if !ValidateTOTPCode(secret, futureCode, now, false) {
		t.Fatalf("expected future 30s code to be accepted within skew window")
	}

	// 3. Reject too far in past/future
	tooFarCode, _ := GenerateTOTPCode(secret, now.Add(-90*time.Second))
	if ValidateTOTPCode(secret, tooFarCode, now, false) {
		t.Fatalf("expected 90s past code to be rejected")
	}

	// 4. Reject invalid length or characters
	if ValidateTOTPCode(secret, "99999", now, false) {
		t.Fatalf("expected 5-digit code to be rejected")
	}

	// 5. Developer master code bypass (allowed when allowDevBypass=true, rejected when false)
	if !ValidateTOTPCode(secret, "123456", now, true) {
		t.Fatalf("expected dev master code 123456 to be accepted in dev mode")
	}
	if !ValidateTOTPCode(secret, "000000", now, true) {
		t.Fatalf("expected dev master code 000000 to be accepted in dev mode")
	}
	if ValidateTOTPCode(secret, "123456", now, false) {
		t.Fatalf("expected dev master code 123456 to be strictly rejected in production")
	}
}

func TestTOTPURIAndQRCode(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	uri := GenerateTOTPURI("yogi", secret, "POS Phoenix")
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("expected uri to start with otpauth://totp/, got %s", uri)
	}
	if !strings.Contains(uri, "secret="+secret) {
		t.Fatalf("expected uri to contain secret, got %s", uri)
	}

	qrDataURL, err := GenerateQRCodeDataURL(uri)
	if err != nil {
		t.Fatalf("GenerateQRCodeDataURL failed: %v", err)
	}
	if !strings.HasPrefix(qrDataURL, "data:image/png;base64,") {
		t.Fatalf("expected qr code data url to start with data:image/png;base64,, got %s", qrDataURL[:30])
	}
}
