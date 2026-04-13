package pusher

// Config holds the configuration for the Pusher client.
type Config struct {
	_       struct{} `json:"-"`
	AppID   string   `env:"APP_ID"  json:"appId"`
	Key     string   `env:"KEY"     json:"key"`
	Secret  string   `env:"SECRET"  json:"secret"`
	Cluster string   `env:"CLUSTER" json:"cluster"`
}

// Enabled returns true if the Pusher config has all required fields set.
func (c *Config) Enabled() bool {
	return c.AppID != "" && c.Key != "" && c.Secret != "" && c.Cluster != ""
}
