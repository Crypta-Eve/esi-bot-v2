package eb2

type Config struct {
	SlackOauthToken    string `envconfig:"BOT_OAUTH_ACCESS_TOKEN" required:"true"`
	SlackBotOauthToken string `envconfig:"BOT_USER_OAUTH_ACCESS_TOKEN"`

	ApiPort uint `envconfig:"API_PORT" default:"5000"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
