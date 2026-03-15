package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// SlackVerifier middleware verifies the signature of incoming requests from Slack.
func SlackVerifier(next http.HandlerFunc) http.HandlerFunc {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	return func(w http.ResponseWriter, r *http.Request) {
		if signingSecret == "" {
			http.Error(w, "Slack signing secret not configured", http.StatusInternalServerError)
			return
		}

		timestamp := r.Header.Get("X-Slack-Request-Timestamp")
		signature := r.Header.Get("X-Slack-Signature")

		if timestamp == "" || signature == "" {
			log.Printf("[VERIFIER] Missing Slack headers. Request rejected.")
			http.Error(w, "Missing Slack headers", http.StatusUnauthorized)
			return
		}

		// Check if the request is older than 5 minutes
		t, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
			return
		}

		if time.Now().Unix()-t > 60*5 {
			http.Error(w, "Request timed out", http.StatusUnauthorized)
			return
		}

		// Read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Could not read body", http.StatusInternalServerError)
			return
		}
		// Restore r.Body so next handler can read it
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Verify signature
		baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
		h := hmac.New(sha256.New, []byte(signingSecret))
		h.Write([]byte(baseString))
		expectedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(h.Sum(nil)))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			log.Printf("[VERIFIER] Invalid signature. Expected %s, got %s", expectedSignature, signature)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		log.Printf("[VERIFIER] Signature verified successfully.")

		next.ServeHTTP(w, r)
	}
}
