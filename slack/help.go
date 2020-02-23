package slack

import (
	"fmt"
	"strings"

	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

func (s *service) makeHelpMessage(event Event) {

	var blob []string

	catLen := len(s.commands)

	for i, category := range s.commands {

		blob = append(blob, fmt.Sprintf("%s - (%s):\n", category.Name, category.Description))

		for _, command := range category.Commands {
			blob = append(blob, fmt.Sprintf("\t%s", command.HelpTextFunc(command)))
		}

		if i != catLen-1 {
			blob = append(blob, "\n")
		}

	}
	text := fmt.Sprintf("```%s```", strings.Join(blob, ""))

	if unknown, ok := event.meta["unknown"].(bool); ok {
		if unknown {
			pretext := "Hmmmm....I'm not sure what you are requesting. Review the options below and try that command again\n"
			text = fmt.Sprintf("%s%s", pretext, text)
		}
	}

	s.logger.Info("Responding to request for help")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(text, false))
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for help.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded with request for help")

}
