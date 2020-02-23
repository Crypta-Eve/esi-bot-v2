package slack

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirkon/go-format"
	"github.com/sirupsen/logrus"

	nslack "github.com/nlopes/slack"
)

func (s *service) makeGHIssueMessage(event Event) {
	if !strings.HasPrefix(event.trigger, "#") {
		return
	}

	event.trigger = strings.TrimPrefix(event.trigger, "#")
	s.logger.Info(event.trigger)

	number, err := strconv.Atoi(event.trigger)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))
	}

	issue, response, err := s.gogithub.Issues.Get(context.Background(), "esi", "esi-issues", number)
	if err != nil {
		_, _, _ = s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(err.Error(), false))
	}
	text := ""
	// https://developer.github.com/v3/issues/#get-a-single-issue
	switch response.StatusCode {
	case http.StatusMovedPermanently, http.StatusNotFound, http.StatusGone:
		text = format.Formatm("${status}", format.Values{
			"status": response.Status,
		})
	case http.StatusOK:
		url := ""
		if issue.HTMLURL != nil {
			url = *issue.HTMLURL
		}
		text = format.Formatm("${status}\n\n${link}", format.Values{
			"status": response.Status,
			"link":   url,
		})
	}

	s.logger.Info("Responding to a request for a gh issue link")
	channel, timestamp, err := s.goslack.PostMessage(event.origin.Channel, nslack.MsgOptionText(text, false), nslack.MsgOptionPostMessageParameters(nslack.PostMessageParameters{
		UnfurlLinks: true,
	}))
	if err != nil {
		s.logger.WithError(err).Error("failed to a request for a gh issue link.")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"channel":   channel,
		"timestamp": timestamp,
	}).Info("successfully responded to a request for a gh issue link")

}
