package eb2

type Config struct {
	ApiPort uint `envconfig:"API_PORT" default:"5000"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}
