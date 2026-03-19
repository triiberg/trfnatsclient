package natsclient

import (
	"fmt"
	"log"

	nats "github.com/nats-io/nats.go"
)

const (
	StreamEmbedding     = "TRF_EMBEDDING"
	SubjectEmbedRequest = "embed.request"

	StreamAudit       = "TRF_AUDIT"
	SubjectAuditEvent = "audit.event"
)

// EmbedRequest is published by domain services when embeddable content changes.
type EmbedRequest struct {
	OrgID       string `json:"org_id"`       // UUID as string
	Domain      string `json:"domain"`       // "journal_entry" | "invoice" | "contact"
	RecordID    string `json:"record_id"`    // UUID as string
	Text        string `json:"text"`
	ContentHash string `json:"content_hash"` // SHA256(text) — skip if unchanged
}

// AuditEvent is published by domain services on every data mutation.
type AuditEvent struct {
	EventID    string `json:"event_id"` // uuid — idempotency key
	OrgID      string `json:"org_id"`
	Service    string `json:"service"`  // "ledger"|"invoices"|"crm"|"payments"
	EntityType   string `json:"entity_type"`  // "journal_entry"|"invoice"|"contact"|"payment"|etc.
	EntityID     string `json:"entity_id"`    // uuid
	Action       string `json:"action"`       // "create"|"update"|"delete"|"post"|"void"|"confirm"|etc.
	ActorID      string `json:"actor_id"`     // account uuid from JWT
	ActorName    string `json:"actor_name"`   // username from JWT
	OccurredAt   string `json:"occurred_at"`  // RFC3339
	Summary      string `json:"summary"`      // e.g. "Invoice INV-001 confirmed"
}

// EnsureStreams creates required JetStream streams if they do not exist.
// Safe to call on every service startup.
func EnsureStreams(js nats.JetStreamContext) error {
	if err := ensureStream(js, StreamEmbedding, []string{SubjectEmbedRequest}); err != nil {
		return err
	}
	return ensureStream(js, StreamAudit, []string{SubjectAuditEvent})
}

func ensureStream(js nats.JetStreamContext, name string, subjects []string) error {
	_, err := js.StreamInfo(name)
	if err == nil {
		log.Printf(`{"op":"nats.stream.ensure.exists","stream":"%s"}`, name)
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("stream info %s: %w", name, err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		MaxAge:   0, // no expiry
	})
	if err != nil {
		return fmt.Errorf("add stream %s: %w", name, err)
	}
	log.Printf(`{"op":"nats.stream.ensure.created","stream":"%s"}`, name)
	return nil
}
