package slack

import (
	"fmt"

	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

func (s *service) makeGreeting(event Event) {

	var head, tail string

	user := event.origin.User
	format := "%s <@%s>%s"
	if len(event.args) == 1 {
		format = "%s %s%s"
		user = event.args[0]
	}

	text := ""

	switch event.trigger {
	case "o7", "o/":
		head = "o7"
		text = fmt.Sprintf(format, head, user, tail)
	case "7o", "\\o":
		head = "7o"
		text = fmt.Sprintf(format, head, user, tail)
	default:
		head = "hey"
		tail = ", hows it goin?"
		text = fmt.Sprintf(format, head, user, tail)
	}

	s.logger.Info("Responding to greeting request")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(text, false))
	if err != nil {
		s.logger.WithError(err).Error("failed to respond to request for help.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded greeting request")

}
