package eb2

type Config struct {
	SlackAPIToken         string `envconfig:"SLACK_API_TOKEN" required:"true"`
	SlackSigningSecret    string `envconfig:"SLACK_SIGNING_SECRET" required:"true"`
	SlackAllowedChannels  string `envconfig:"SLACK_ALLOWED_CHANNELS" required:"true"`
	SlackPrefix           string `envconfig:"SLACK_PREFIX" required:"true"`
	SlackESIStatusChannel string `envconfig:"SLACK_ESISTATUS_CHANNEL" required:"true"`
	SlackSendStartupMsg   bool   `envconfig:"SLACK_SEND_STARTUP_MSG" default:"true"`
	SlackLegacyAPIToken   string `envconfig:"SLACK_LEGACY_API_TOKEN" required:"true"`
	SlackModChannel       string `envconfig:"SLACK_MOD_CHANNEL" required:"true"`

	EveClientID     string `envconfig:"EVE_CLIENT_ID" required:"true"`
	EveClientSecret string `envconfig:"EVE_CLIENT_SECRET" required:"true"`
	EveCallback     string `envconfig:"EVE_CALLBACK" required:"true"`

	ApiPort uint `envconfig:"API_PORT" default:"5000"`

	AppVersion string `envconfig:"APP_VERSION" required:"true"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
