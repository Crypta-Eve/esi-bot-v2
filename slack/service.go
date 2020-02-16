package slack

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type Service interface {
	HandleSlashCommand(ctx context.Context, parsed url.Values)
}

type service struct {
	logger *logrus.Logger
}

type response struct {
	DeleteOriginal  bool   `json:"delete_original,omitempty"`
	ReplaceOriginal bool   `json:"replace_original,omitempty"`
	UnfurlLinks     bool   `json:"unfurl_links,omitempty"`
	ResponseType    string `json:"response_type,omitempty"`
	Text            string `json:"text"`
}

func New(logger *logrus.Logger) Service {
	return &service{
		logger: logger,
	}
}

func (s *service) HandleSlashCommand(ctx context.Context, parsed url.Values) {

	var (
		message []byte
		err     error
	)

	time.Sleep(time.Millisecond * 500)

	command := parsed.Get("text")
	responseURL := parsed.Get("response_url")
	// Check for an empty command or if they are asking for help
	if command == "" || command == "help" {
		if responseURL == "" {
			s.logger.WithFields(logrus.Fields{
				"user_id":    parsed.Get("user_id"),
				"user_name":  parsed.Get("user_name"),
				"team_id":    parsed.Get("team_id"),
				"channel_id": parsed.Get("channel_id"),
				"text":       parsed.Get("text"),
			}).Error("unable to respond due to missing response url")
		}
		message, err = s.makeHelpMessage(true)
		if err != nil {
			s.logger.WithError(err).Error("failed to marshal response to command")
			return
		}
		err = s.sendSlackResponse(responseURL, message)
		if err != nil {
			s.logger.WithError(err).Error("failed to reply to command")
			return
		}
		return
	}

	switch {
	case strSliceContainsString(greetings, command):
		message, err = s.makeHelloMessage(parsed.Get("user_id"), parsed.Get("text"))
	case strSliceContainsString(links, command):
		message, err = s.makeLinkMessage(command)
	default:
		message, err = s.makeHelpMessage(false)
	}

	if err != nil {
		s.logger.WithError(err).Error("failed to reply to command")
		return
	}

	err = s.sendSlackResponse(responseURL, message)
	if err != nil {
		s.logger.WithError(err).Error("failed to reply to command")
		return
	}
}

func (s *service) sendSlackResponse(url string, data []byte) error {
	_, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	return err
}

func strSliceContainsString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
