package slack

import (
	"context"
	"strings"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/google/go-github/v29/github"
	nslack "github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/sirupsen/logrus"
)

type Service interface {
	HandleMessageEvent(context.Context, *slackevents.MessageEvent)
}

type service struct {
	logger   *logrus.Logger
	config   *eb2.Config
	commands []Category
	flat     []Command
	channels []string
	goslack  *nslack.Client
	gogithub *github.Client
}

var (
	layoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
	rl        = [][]string{}
)

func New(logger *logrus.Logger, config *eb2.Config) Service {

	s := &service{
		logger:   logger,
		config:   config,
		channels: strings.Split(config.SlackAllowedChannels, ","),
		goslack:  nslack.New(config.SlackAPIToken),
		gogithub: github.NewClient(nil),
	}

	commands := s.BuildCommands()

	s.commands = commands
	s.flat = s.flattenCommands(commands)

	routes, err := fetchRouteStatuses("latest")
	if err != nil {
		logger.WithError(err).Fatal("failed to load esi route status")
	}
	for _, route := range routes {
		if route.Method == "get" {
			s := strings.TrimPrefix(route.Route, "/")
			s = strings.TrimSuffix(s, "/")
			rl = append(rl, strings.Split(s, "/"))
		}
	}

	go func(channels []string, text string) {
		for _, c := range channels {
			_, _, _ = s.goslack.PostMessage(c, nslack.MsgOptionText(getStartupMessage(), false))
			time.Sleep(time.Millisecond * 500)
		}
	}(s.channels, getStartupMessage())

	return s

}

func (s *service) flattenCommands(commands []Category) []Command {
	var list = make([]Command, 0)
	for _, cat := range commands {
		for _, com := range cat.Commands {
			com.Category = cat
			list = append(list, com)
		}
	}

	return list
}

func (s *service) HandleMessageEvent(ctx context.Context, sevent *slackevents.MessageEvent) {

	if !strInStrSlice(sevent.Channel, s.channels) {
		return
	}

	if !strings.HasPrefix(sevent.Text, s.config.SlackPrefix) {
		return
	}

	// Split up the text of the Event into a slice
	// Example: !esi status = []string{"!esi", "status"}
	// Example: !esi types 32 = []string{"!esi", "types", "32"}
	// Example: !esi issues = []string{"!esi", "issues"}
	// Example: !esi status --location=china = []string{"!esi", "status", "--location=china"}
	text := strings.Split(sevent.Text, " ")
	// If text contains just the bot prefix, then we should reply with help.
	if len(text) == 1 {
		s.makeHelpMessage(Event{
			origin:  sevent,
			trigger: text[0],
		})
		return
	}
	// At this stage with have a slice that has a length >= 2, with the first entry still being the bot prefix
	// This is predictable because we checked to see if the message started with the prefix earlier and have since then only split the text into a slice on spaces
	// Now we want to drop the prefix. It no longer serves a purpose
	text = text[1:]

	// The command that is to be invoked should have a single word trigger and that trigger should now be at index 0 of the text slice.
	// Lets use the string at index 0 to determine which command is being invoked.
	// If missing is true, that means that the text at index 0 of the text slice did prompt any of the trigger funcs to return true
	// We do not know which command is being trigger, return the help prompt with a small message stating that we could not match the request
	// to a known command
	invokedCommand, missing := s.determineInvokedCommand(text[0])
	if missing {
		s.makeHelpMessage(Event{
			origin: sevent,
			meta: map[string]interface{}{
				"unknown": true,
			},
		})
		return
	}
	// Construct Internal Event to hold trigger, args, and flags
	event := Event{
		origin:  sevent,
		trigger: text[0],
	}

	// Determine if the rest of the text is long enough to be an arg or flag
	if len(text) >= 1 {
		event.flags = make(map[string]string)
		// Treat remain pieces as args until we can parse them into flags
		for _, arg := range text[1:] {
			if strings.HasPrefix(arg, "--") {
				arg = strings.TrimPrefix(arg, "--")
				slFlag := strings.Split(arg, "=")
				if len(slFlag) != 2 {
					continue
				}

				event.flags[slFlag[0]] = slFlag[1]
				continue
			}

			// If the arg does not start with a --, then treat it as an actual arguement.
			// The invoked command will be able to deal with these if they want to
			event.args = append(event.args, arg)
		}
	}

	// We now have the invoked command. Here we call that command. Our job as the entry point is done
	// The invoke command will either error out, send a message with the requested message to slack or it
	// will send a message with an error message to slack assisting the user with sending the correct input
	invokedCommand.Action(event)

}

func (s *service) determineInvokedCommand(command string) (*Command, bool) {

	var triggered *Command

	for _, com := range s.flat {

		if com.TriggerFunc(com, command) {
			triggered = &com
			break
		}
	}
	return triggered, triggered == nil

}

// Takes a str and a slice of str and tell you if the str is in the slice
func strInStrSlice(needle string, haystack []string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
