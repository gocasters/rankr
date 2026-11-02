package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

const (
	testHookID      = "12345"
	testEventName   = "push"
	testDeliveryUID = "delivery-123"
)

func TestVerifyEventAcceptsValidGitHubSignature(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"ref":"refs/heads/main"}`)
	service := New(nil, nil, nil, secret, 1<<20, "", 0)

	got, err := service.VerifyEvent(strings.NewReader(string(payload)), testHookID, testEventName, testDeliveryUID, signPayload(secret, payload))
	if err != nil {
		t.Fatalf("VerifyEvent() error = %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("VerifyEvent() payload = %q, want %q", got, payload)
	}
}

func TestVerifyEventRejectsMissingGitHubHeaders(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action":"opened"}`)
	signature := signPayload(secret, payload)

	tests := []struct {
		name        string
		hookID      string
		eventName   string
		deliveryUID string
		signature   string
		wantError   string
	}{
		{name: "hook ID", eventName: testEventName, deliveryUID: testDeliveryUID, signature: signature, wantError: "X-GitHub-Hook-ID"},
		{name: "event name", hookID: testHookID, deliveryUID: testDeliveryUID, signature: signature, wantError: "X-GitHub-Event"},
		{name: "delivery ID", hookID: testHookID, eventName: testEventName, signature: signature, wantError: "X-GitHub-Delivery"},
		{name: "signature", hookID: testHookID, eventName: testEventName, deliveryUID: testDeliveryUID, wantError: "X-Hub-Signature-256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(nil, nil, nil, secret, 1<<20, "", 0)
			_, err := service.VerifyEvent(strings.NewReader(string(payload)), tt.hookID, tt.eventName, tt.deliveryUID, tt.signature)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("VerifyEvent() error = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestVerifyEventRejectsMalformedGitHubSignature(t *testing.T) {
	service := New(nil, nil, nil, []byte("test-secret"), 1<<20, "", 0)

	for _, signature := range []string{"deadbeef", "sha256=not-hex"} {
		t.Run(signature, func(t *testing.T) {
			_, err := service.VerifyEvent(strings.NewReader(`{}`), testHookID, testEventName, testDeliveryUID, signature)
			if err == nil {
				t.Fatal("VerifyEvent() error = nil, want malformed signature error")
			}
		})
	}
}

func TestVerifyEventRejectsIncorrectGitHubSignature(t *testing.T) {
	secret := []byte("test-secret")
	service := New(nil, nil, nil, secret, 1<<20, "", 0)
	signature := signPayload(secret, []byte(`{"action":"closed"}`))

	_, err := service.VerifyEvent(strings.NewReader(`{"action":"opened"}`), testHookID, testEventName, testDeliveryUID, signature)
	if err == nil || !strings.Contains(err.Error(), "verification failed") {
		t.Fatalf("VerifyEvent() error = %v, want verification failure", err)
	}
}

func TestVerifyEventRejectsEmptySecret(t *testing.T) {
	payload := []byte(`{"action":"opened"}`)
	service := New(nil, nil, nil, nil, 1<<20, "", 0)

	_, err := service.VerifyEvent(strings.NewReader(string(payload)), testHookID, testEventName, testDeliveryUID, signPayload(nil, payload))
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("VerifyEvent() error = %v, want empty secret error", err)
	}
}

func TestVerifyEventRejectsPayloadOverLimit(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte("12345")
	service := New(nil, nil, nil, secret, 4, "", 0)

	// Signing the truncated prefix proves an oversized body cannot be accepted
	// merely because io.LimitReader silently stopped at the configured limit.
	_, err := service.VerifyEvent(strings.NewReader(string(payload)), testHookID, testEventName, testDeliveryUID, signPayload(secret, payload[:4]))
	if err == nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("VerifyEvent() error = %v, want payload limit error", err)
	}
}

func signPayload(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
