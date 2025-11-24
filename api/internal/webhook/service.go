package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Service dispatches webhook events to subscribed URLs.
type Service struct {
	Repo       Repository
	Secret     string
	TimeoutSec int
	Retries    int
}

func (s *Service) Dispatch(event string, data any) {
	hooks, err := s.Repo.List()
	if err != nil {
		log.Error().
			Err(err).
			Str("event", event).
			Msg("Failed to list webhooks for event dispatch")
		return
	}

	payload := EventPayload{
		Event: event,
		Data:  data,
		Time:  time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("event", event).
			Msg("Failed to marshal webhook payload")
		return
	}

	dispatchCount := 0
	for _, h := range hooks {
		if !h.Enabled || !contains(h.Events, event) {
			continue
		}
		dispatchCount++
		go s.deliver(h.URL, body, event)
	}

	if dispatchCount > 0 {
		log.Info().
			Str("event", event).
			Int("webhook_count", dispatchCount).
			Msg("Dispatching webhook event")
	} else {
		log.Debug().
			Str("event", event).
			Msg("No webhooks configured for event")
	}
}

func (s *Service) deliver(url string, body []byte, event string) {
	timeout := time.Duration(s.TimeoutSec) * time.Second
	retries := s.Retries
	if retries < 0 {
		retries = 0
	}

	log.Debug().
		Str("url", url).
		Str("event", event).
		Int("max_retries", retries).
		Msg("Starting webhook delivery")

	for attempt := 0; attempt <= retries; attempt++ {
		req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if s.Secret != "" {
			req.Header.Set("X-Webhook-Signature", hmacHex([]byte(s.Secret), body))
		}

		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Info().
				Str("url", url).
				Str("event", event).
				Int("status", resp.StatusCode).
				Int("attempt", attempt+1).
				Msg("Webhook delivered successfully")
			_ = resp.Body.Close()
			return
		}

		// Log the failure
		if err != nil {
			log.Warn().
				Err(err).
				Str("url", url).
				Str("event", event).
				Int("attempt", attempt+1).
				Int("max_attempts", retries+1).
				Msg("Webhook delivery failed with error")
		} else {
			log.Warn().
				Str("url", url).
				Str("event", event).
				Int("status", resp.StatusCode).
				Int("attempt", attempt+1).
				Int("max_attempts", retries+1).
				Msg("Webhook delivery failed with non-2xx status")
			_ = resp.Body.Close()
		}

		if resp != nil {
			_ = resp.Body.Close()
		}

		// Wait before retry (exponential backoff)
		if attempt < retries {
			backoff := time.Duration(attempt+1) * 500 * time.Millisecond
			log.Debug().
				Str("url", url).
				Dur("backoff_ms", backoff).
				Msg("Waiting before webhook retry")
			time.Sleep(backoff)
		}
	}

	log.Error().
		Str("url", url).
		Str("event", event).
		Int("attempts", retries+1).
		Msg("Webhook delivery failed after all retries")
}

func hmacHex(secret, data []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
