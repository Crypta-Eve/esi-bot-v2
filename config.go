package eb2

type Config struct {
	SlackAPIToken         string   `envconfig:"SLACK_API_TOKEN" required:"true"`
	SlackSigningSecret    string   `envconfig:"SLACK_SIGNING_SECRET" required:"true"`
	SlackPingChannels     []string `envconfig:"SLACK_PING_CHANNELS" required:"true" split_words:"true"`
	SlackPrefixes         []string `envconfig:"SLACK_PREFIXES" required:"true" split_words:"true"`
	SlackSendStartupMsg   bool     `envconfig:"SLACK_SEND_STARTUP_MSG" default:"true"`
	SlackModChannel       string   `envconfig:"SLACK_MOD_CHANNEL" required:"true"`
	SlackESIChannel       string   `envconfig:"SLACK_ESI_CHANNEL" required:"true"`
	SlackESIStatusChannel string   `envconfig:"SLACK_ESISTATUS_CHANNEL" required:"true"`

	EveClientID     string `envconfig:"EVE_CLIENT_ID" required:"true"`
	EveClientSecret string `envconfig:"EVE_CLIENT_SECRET" required:"true"`
	EveCallback     string `envconfig:"EVE_CALLBACK" required:"true"`

	ApiPort uint `envconfig:"API_PORT" default:"5000"`

	AppVersion string `envconfig:"APP_VERSION" required:"true"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
