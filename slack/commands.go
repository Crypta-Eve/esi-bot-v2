package slack

type Action func(ParsedCommand) (response, error)

type Category struct {
	Name        string
	Description string
	Commands    []Command
	Active      bool
}

type Command struct {
	Name        string
	Description string
	Category    string
	Triggers    []string
	Strict      bool
	Prefix      string
	Action      Action
	Example     string
}

type ParsedCommand struct {
	ChannelID   string `mapstructure:"channel_id"`
	ChannelName string `mapstructure:"channel_name"`
	Command     string `mapstructure:"command"`
	ResponseURL string `mapstructure:"response_url"`
	TeamDomain  string `mapstructure:"team_domain"`
	TeamID      string `mapstructure:"team_id"`
	Text        string `mapstructure:"text"`
	Token       string `mapstructure:"token"`
	TriggerID   string `mapstructure:"trigger_id"`
	UserID      string `mapstructure:"user_id"`
	UserName    string `mapstructure:"user_name"`
}

var commands = []Category{
	Category{
		Name:        "Greetings",
		Description: "Have the bot say hello to you by invoking any of these commands",
		Commands: []Command{
			Command{
				Name:        "hello",
				Description: "Reply with a nice message greeting the user",
				Triggers:    []string{"hello", "hi", "hey", "o7", "o/", "7o", "\\o"},
				Strict:      true,
				Action:      s.makeGreeting,
			},
		},
	},
	Category{
		Name:        "Helpful Links",
		Description: "Forgot a link to the issues board, maybe the docs? ESI Bot can help with that",
		Commands: []Command{
			Command{
				Name:        "esi-issues",
				Description: "Responds with a link to the esi-issues issues board on Github",
				Triggers:    []string{"issues"},
				Strict:      true,
				Action:      s.makeLinkMessage,
			},
			Command{
				Name:        "sso-issues",
				Description: "Responds with a link to the sso-issues issues board on Github",
				Triggers:    []string{"sso"},
				Strict:      true,
				Action:      s.makeLinkMessage,
			},
			Command{
				Name:        "ids",
				Description: "Links to documnetation of various id ranges that one will encounter while using ESI",
				Triggers:    []string{"id", "ids", "ranges"},
				Strict:      true,
				Action:      s.makeLinkMessage,
			},
			Command{
				Name:        "source code",
				Description: "Link to the open source code for ESI Bot",
				Triggers:    []string{"source", "repo"},
				Strict:      true,
				Action:      s.makeLinkMessage,
			},
			Command{
				Name:        "faq",
				Description: "Link to Frequently Asked Questions on ESI Docs Site",
				Triggers:    []string{"faq"},
				Strict:      true,
				Action:      s.makeLinkMessage,
			},
		},
	},
}
