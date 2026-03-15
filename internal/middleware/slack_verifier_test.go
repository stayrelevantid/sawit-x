package middleware_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/indragiri/sawit-x/internal/middleware"
)

const testSigningSecret = "test_slack_signing_secret"

// buildSlackRequest creates a fake Slack HTTP request with a valid HMAC signature.
func buildSlackRequest(t *testing.T, body string, secret string, tsOverride ...int64) *http.Request {
	t.Helper()
	ts := time.Now().Unix()
	if len(tsOverride) > 0 {
		ts = tsOverride[0]
	}

	baseString := fmt.Sprintf("v0:%d:%s", ts, body)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(baseString))
	sig := fmt.Sprintf("v0=%s", hex.EncodeToString(h.Sum(nil)))

	req := httptest.NewRequest(http.MethodPost, "/slack/events", bytes.NewBufferString(body))
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(ts, 10))
	req.Header.Set("X-Slack-Signature", sig)
	return req
}

func TestSlackVerifier_ValidRequest(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", testSigningSecret)

	called := false
	handler := middleware.SlackVerifier(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Verify body is still readable after middleware
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("could not read body after middleware: %v", err)
		}
		if string(body) != "test_body" {
			t.Errorf("expected body 'test_body', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusOK)
	})

	req := buildSlackRequest(t, "test_body", testSigningSecret)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !called {
		t.Error("expected inner handler to be called, but it was not")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestSlackVerifier_InvalidSignature(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", testSigningSecret)

	handler := middleware.SlackVerifier(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should NOT be called with invalid signature")
	})

	req := buildSlackRequest(t, "test_body", "wrong_secret")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestSlackVerifier_ExpiredTimestamp(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", testSigningSecret)

	handler := middleware.SlackVerifier(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should NOT be called with expired timestamp")
	})

	// 6 minutes ago = expired
	oldTS := time.Now().Unix() - 360
	req := buildSlackRequest(t, "test_body", testSigningSecret, oldTS)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestSlackVerifier_MissingHeaders(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", testSigningSecret)

	handler := middleware.SlackVerifier(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should NOT be called without headers")
	})

	req := httptest.NewRequest(http.MethodPost, "/slack/events", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}
