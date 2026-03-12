package natsclient

import (
	"fmt"
	"log"

	nats "github.com/nats-io/nats.go"
)

const (
	StreamEmbedding     = "TRF_EMBEDDING"
	SubjectEmbedRequest = "embed.request"
)

// EmbedRequest is published by domain services when embeddable content changes.
type EmbedRequest struct {
	TenantSchema string `json:"tenant_schema"`
	OrgID        string `json:"org_id"`    // UUID as string
	Domain       string `json:"domain"`    // "journal_entry" | "invoice" | "contact"
	RecordID     string `json:"record_id"` // UUID as string
	Text         string `json:"text"`
	ContentHash  string `json:"content_hash"` // SHA256(text) — skip if unchanged
}

// EnsureStreams creates required JetStream streams if they do not exist.
// Safe to call on every service startup.
func EnsureStreams(js nats.JetStreamContext) error {
	_, err := js.StreamInfo(StreamEmbedding)
	if err == nil {
		log.Printf(`{"op":"nats.stream.ensure.exists","stream":"%s"}`, StreamEmbedding)
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("stream info: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     StreamEmbedding,
		Subjects: []string{SubjectEmbedRequest},
		Storage:  nats.FileStorage,
		MaxAge:   0, // no expiry
	})
	if err != nil {
		return fmt.Errorf("add stream: %w", err)
	}
	log.Printf(`{"op":"nats.stream.ensure.created","stream":"%s"}`, StreamEmbedding)
	return nil
}
