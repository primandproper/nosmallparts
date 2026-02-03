package webhooks

type (
	UserDataCollection struct {
		Data map[string][]Webhook `json:"data,omitempty"`
	}
)
