package tracking

// EventPayload is one event as sent by the client.
type EventPayload struct {
	DeviceID   string         `json:"device_id"`
	Event      string         `json:"event"`
	Properties map[string]any `json:"properties"`
	Referrer   string         `json:"referrer"`
}

// IngestRequest is the envelope accepted by POST /api/track.
type IngestRequest struct {
	Events []EventPayload `json:"events" binding:"required,min=1,max=50,dive"`
}
