package eb2

type Config struct {
	SlackAPIToken string `envconfig:"SLACK_API_TOKEN" required:"true"`

	ApiPort uint `envconfig:"API_PORT" default:"5000"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
