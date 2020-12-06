package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/google/go-github/v29/github"
	"github.com/nlopes/slack"
	nslack "github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run()
	ProcessEvent(context.Context, *slackevents.MessageEvent)
}

type service struct {
	logger   *logrus.Logger
	config   *eb2.Config
	commands []Category
	flat     []Command
	goslack  *nslack.Client
	gogithub *github.Client
	client   *http.Client
	caches   map[string]*cache.Cache
}

var (
	layoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
	rl        = [][]string{}
)

func New(logger *logrus.Logger, config *eb2.Config) Service {

	s := &service{
		logger:   logger,
		config:   config,
		goslack:  nslack.New(config.SlackAPIToken),
		gogithub: github.NewClient(nil),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		caches: map[string]*cache.Cache{
			"routes": cache.New(cache.NoExpiration, cache.NoExpiration),
			"etags":  cache.New(cache.NoExpiration, cache.NoExpiration),
		},
	}

	commands := s.BuildCommands()

	s.commands = commands
	s.flat = s.flattenCommands(commands)
	version := "latest"
	routes, err := s.fetchRouteStatuses(version)
	if err != nil {
		logger.WithError(err).Fatal("failed to load esi route status")
	}

	if routes != nil {
		s.caches["routes"].Set(version, routes, 0)
	}

	for _, route := range routes {
		if route.Method == "get" {
			s := strings.TrimPrefix(route.Route, "/")
			s = strings.TrimSuffix(s, "/")
			rl = append(rl, strings.Split(s, "/"))
		}
	}

	if config.SlackSendStartupMsg {
		go func(channels []string) {
			for _, c := range channels {
				_, _, _ = s.goslack.PostMessage(c, nslack.MsgOptionText(getStartupMessage(), false))
				time.Sleep(time.Millisecond * 100)
			}
		}(config.SlackPingChannels)
	}
	return s

}

func (s *service) Run() {

	version := "latest"
	var cachedRoutes []*eb2.ESIStatus
	var found bool
	cachedRoutes, found = s.checkCache(version)
	if !found {
		routes, err := s.fetchRouteStatuses(version)
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch routes statuses from ESI")
			return
		}

		if routes == nil {
			return
		}

		s.caches["routes"].Set(version, routes, time.Minute*2)

		// return from the func early
		return
	}

	updatedRoutes, err := s.fetchRouteStatuses(version)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch route statuses")
		return
	}
	if updatedRoutes == nil {
		s.logger.Info("no change is routes detected, exiting early")
		return
	}

	countUpdatedRoutes := len(updatedRoutes)
	countCachedRoutes := len(cachedRoutes)

	if countUpdatedRoutes != countCachedRoutes {
		mutatedRoutes := []string{}

		// The number of routes that ESI has has changed. This can be because a route has been removed or a route has been added. Lets try to figure that out
		if countUpdatedRoutes > countCachedRoutes {
			// A route has been added to ESI, lets try to figure out which route(s) it was
			for _, route := range updatedRoutes {
				var found = false
				for _, cachedRoute := range cachedRoutes {
					if route.Route == cachedRoute.Route {
						found = true
						break
					}
				}
				if !found {
					mutatedRoutes = append(mutatedRoutes, fmt.Sprintf("+ %s", route.Route))
				}
			}
		} else if countUpdatedRoutes < countCachedRoutes {
			// A route has been removed from ESI, lets try to figure out which route(s) it was
			for _, route := range cachedRoutes {
				var found = false
				for _, updatedRoute := range updatedRoutes {
					if route.Route == updatedRoute.Route {
						found = true
						break
					}
				}
				if !found {
					mutatedRoutes = append(mutatedRoutes, fmt.Sprintf("- %s", route.Route))
				}
			}
		}

		s.MakeESIMutatedRoutesMessage(s.config.SlackESIChannel, mutatedRoutes)

	}

	s.MakeESIStatusMessage(s.config.SlackESIStatusChannel, updatedRoutes, "latest")

}

func (s *service) ProcessEvent(ctx context.Context, sevent *slackevents.MessageEvent) {

	valid := false
	for _, prefix := range s.config.SlackPrefixes {
		if strings.HasPrefix(sevent.Text, prefix) {
			valid = true
			break
		}
	}

	for _, specialTerm := range []string{"xml", "crest"} {

		if strings.Contains(sevent.Text, specialTerm) {
			err := s.goslack.AddReaction("rip", slack.NewRefToMessage(sevent.Channel, sevent.TimeStamp))
			if err != nil {
				s.logger.WithError(err).Error("failed to reaction message with legacy term")
			}
			break
		}

	}

	if !valid {
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
