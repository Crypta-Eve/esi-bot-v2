package slack

type Action func(SlashCommand) (Msg, error)

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
	Args        map[string][]string
	Strict      bool
	Prefix      string
	Action      Action
	Example     string
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
	Category{
		Name:        "ESI Issues",
		Description: "Helpful links for communicating bugs, feature requests, and inconsistencies with the ESI API to the ESI Developers",
		Commands: []Command{
			Command{
				Name:        "new issue",
				Description: "Instructions for opening a new ESI issue",
				Triggers:    []string{"new"},
				Strict:      true,
				Action:      s.makeIssues,
			},
			Command{
				Name:        "bug report",
				Description: "Instructions for reporting an ESI bug",
				Triggers:    []string{"br", "bug"},
				Strict:      true,
				Action:      s.makeIssues,
			},
			Command{
				Name:        "feature request",
				Description: "Return instructions for creating a new feature request",
				Triggers:    []string{"fr", "feature", "enhancement"},
				Strict:      true,
				Action:      s.makeIssues,
			},
			Command{
				Name:        "inconsistency",
				Description: "Return instructions for reporting an inconsistency.",
				Triggers:    []string{"incon", "inconsistency"},
				Strict:      true,
				Action:      s.makeIssues,
			},
		},
	},
	Category{
		Name:        "status",
		Description: "Commands that can be called to check on the status of endpoint on the ESI API",
		Commands: []Command{
			Command{
				Name:        "status",
				Description: "Check the health of the endpoints on the ESI API",
				Triggers:    []string{"status"},
				Args: map[string][]string{
					"version": []string{
						"meta", "legacy", "dev", "latest",
					},
				},
				Strict: true,
				Action: s.makeESIStatusMessage,
			},
			Command{
				Name:        "server status",
				Description: "Check the uptime and player count of the Eve Servers",
				Triggers:    []string{"tq", "tranquility"},
				Args:        map[string][]string{},
				Strict:      true,
				Action:      s.makeEveServerStatusMessage,
			},
		},
	},
}
