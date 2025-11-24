package webhook

// Webhook is stored in DB.
type Webhook struct {
	ID      int64
	URL     string
	Events  []string
	Enabled bool
}

// WebhookDTO is sent/received over the API.
type WebhookDTO struct {
	ID      int64    `json:"id" example:"1" doc:"Webhook ID"`
	URL     string   `json:"url" example:"https://example.com/webhook" doc:"Webhook endpoint URL"`
	Events  []string `json:"events" example:"firmware.uploaded,firmware.deleted" doc:"Events to subscribe to"`
	Enabled bool     `json:"enabled" example:"true" doc:"Whether webhook is active"`
}

type EventPayload struct {
	Event string `json:"event" example:"firmware.uploaded" doc:"Event type"`
	Data  any    `json:"data" doc:"Event-specific payload data"`
	Time  string `json:"time" example:"2024-01-15T10:30:00Z" doc:"Event timestamp in RFC3339 format"`
}
