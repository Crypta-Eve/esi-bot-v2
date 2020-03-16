package slack

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eveisesi/eb2/tools"
	"github.com/sirkon/go-format"
	"github.com/sirupsen/logrus"

	nslack "github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

type Action func(Event)
type TriggerFunc func(Command, string) bool
type HelpTextFunc func(Command) string
type ExampleGen func(Command) string

type Event struct {
	origin  *slackevents.MessageEvent
	trigger string
	args    []string
	flags   map[string]string
	meta    map[string]interface{}
}

type Flags map[string][]string

func (x Flags) HasFlag(s string) bool {
	if _, err := x[s]; err {
		return false
	}

	return true
}

func (x Flags) IsValidValue(s, v string) bool {
	if !x.HasFlag(s) {
		return false
	}

	m := x[s]
	if len(m) == 0 {
		return false
	}

	for _, a := range m {
		if a == v {
			return true
		}
	}

	return false
}

type Category struct {
	Name        string
	Description string
	Commands    []Command
	Active      bool
}

type Command struct {
	Description  string
	Category     Category
	TriggerFunc  TriggerFunc
	Args         []string
	Flags        Flags
	Action       Action
	HelpTextFunc HelpTextFunc
	triggers     []string
	example      ExampleGen
}

func (s *service) BuildCommands() []Category {
	return []Category{
		Category{
			Name:        "Help",
			Description: "Get Help Using ESI Bot V2",
			Commands: []Command{
				Command{
					Description: "Reply with the help dialog",
					TriggerFunc: func(c Command, s string) bool {
						return s == "" || s == "help"
					},
					Action: s.makeHelpMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"help"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Get the current version of the application",
					TriggerFunc: func(c Command, s string) bool {
						return s == "version"
					},
					Action: func(event Event) {

						text := fmt.Sprintf("Current Version: %s", s.config.AppVersion)

						s.logger.Info("Responding to request for help (ephemeral)")
						channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(text, false))
						if err != nil {
							s.logger.WithError(err).Error("failed to respond to request for help (ephemeral).")
							return
						}
						s.logger.WithFields(logrus.Fields{
							"user":      event.origin.User,
							"channel":   channel,
							"timestamp": timestamp,
						}).Info("successfully responded with request for help (ephemeral).")

					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"version"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
			},
		},
		Category{
			Name:        "Greetings",
			Description: "Have the bot say hello to you by invoking any of these commands",
			Commands: []Command{
				Command{
					Description: "Reply with a nice message greeting the user",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeGreeting,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"hello", "hi", "hey", "o7", "o/", "7o", `\o`},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
			},
		},
		Category{
			Name:        "Helpful Links",
			Description: "Forgot a link to the issues board, maybe the docs? ESI Bot can help with that",
			Commands: []Command{
				Command{
					Description: "Responds with a link to the esi-issues issues board on Github",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"issues"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Responds with a link to the sso-issues issues board on Github",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"sso"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Links to documnetation of various id ranges that one will encounter while using ESI",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)

					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"id", "ids", "ranges"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Link to the open source code for ESI Bot",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"source", "repo"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Link to Frequently Asked Questions on ESI Docs Site",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"faq"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Responds with a link to the ESI Swagger UI",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"webui", "ui"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Responds with a link to the ESI API Diff",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					Action: s.makeLinkMessage,
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					triggers: []string{"diff", "diffs"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
			},
		},
		Category{
			Name:        "ESI Issues",
			Description: "Helpful links for communicating bugs, feature requests, and inconsistencies with the ESI API to the ESI Developers",
			Commands: []Command{
				Command{
					Description: "Instructions for opening a new ESI issue",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					triggers: []string{"new"},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeIssues,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Instructions for reporting an ESI bug",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					triggers: []string{"br", "bug"},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeIssues,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Return instructions for creating a new feature request",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					triggers: []string{"fr", "feature", "enhancement"},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeIssues,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Return instructions for reporting an inconsistency.",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					triggers: []string{"incon", "inconsistency"},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeIssues,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
			},
		},
		Category{
			Name:        "EVE Status",
			Description: "Commands that can be called to check on the status of endpoint on the ESI API",
			Commands: []Command{
				Command{
					Description: "Check the health of the endpoints on the ESI API",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Flags: map[string][]string{
						"version": []string{
							"meta", "legacy", "dev", "latest",
						},
					},
					Action:   s.handleESIStatusMessage,
					triggers: []string{"status"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Check the uptime and player count of the Eve Servers",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action:   s.handleEveTQStatus,
					triggers: []string{"tq", "tranquility"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Check the uptime and player count of the Serenity Eve Servers",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action:   s.handleEveSerenityStatus,
					triggers: []string{"serenity", "china"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
			},
		},
		Category{
			Name:        "Requests",
			Description: "Commands that allow you to make requests to supported external APIs including ESI",
			Commands: []Command{
				Command{
					Description: "Use this command to easily call out to /universe/types/:id",
					TriggerFunc: func(c Command, s string) bool {
						return strInStrSlice(s, c.triggers)
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     strings.Join(c.triggers, ", "),
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action:   s.makeESITypeRequestMessage,
					triggers: []string{"item", "item_id", "type", "type_id"},
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": c.triggers[tools.UnsignedRandomIntWithMax(len(c.triggers)-1)],
						})
					},
				},
				Command{
					Description: "Any string begining with a `/` followed by a valid version number will trigger a request to ESI",
					TriggerFunc: func(c Command, s string) bool {
						return strings.HasPrefix(s, "/")
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     "/{latest, legacy, dev, v1, v2, v3, v4, v5, v6}/...",
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeESIDynamicRequestMessage,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": "/latest/universe/types",
						})
					},
				},
				Command{
					Description: "Any string begining with a `#` followed an integer will trigger a look up of that issue on Github",
					TriggerFunc: func(c Command, s string) bool {
						re, _ := regexp.Compile("^#[0-9]+")
						return re.Match([]byte(s))
					},
					HelpTextFunc: func(c Command) string {
						return format.Formatm("${trigger}\n\t${description} (i.e. ${example})\n", format.Values{
							"trigger":     "#[0-9]",
							"description": c.Description,
							"example":     c.example(c),
						})
					},
					Action: s.makeGHIssueMessage,
					example: func(c Command) string {
						return format.Formatm("${prefix} ${trigger}", format.Values{
							"prefix":  s.config.SlackPrefix,
							"trigger": "/latest/universe/types",
						})
					},
				},
			},
		},
	}
}
