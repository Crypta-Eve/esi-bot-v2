package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Service interface {
	HandleSlashCommand(context.Context, ParsedCommand)
}

type service struct {
	logger   *logrus.Logger
	commands []Category
	flat     []Command
}

type response struct {
	DeleteOriginal  bool   `json:"delete_original,omitempty"`
	ReplaceOriginal bool   `json:"replace_original,omitempty"`
	UnfurlLinks     bool   `json:"unfurl_links,omitempty"`
	ResponseType    string `json:"response_type,omitempty"`
	Text            string `json:"text"`
}

func (s *service) flattenCommands() {
	for _, cat := range s.commands {
		for _, com := range cat.Commands {

			if !com.Strict && com.Prefix == "" {
				s.logger.Panicf("Non Strict Command with empty prefix detected. Prefix is required on non strict commands")
				os.Exit(1)
			}

			s.flat = append(s.flat, com)
		}
	}
}

var s = &service{}

func New(logger *logrus.Logger) Service {
	s.commands = commands
	s.logger = logger
	s.flattenCommands()
	return s

}

var (
	res     response
	message []byte
	err     error
)

func (s *service) HandleSlashCommand(ctx context.Context, parsed ParsedCommand) {
	time.Sleep(time.Millisecond * 250)

	// Check to see if this is a call for help
	if parsed.Text == "" || parsed.Text == "help" {

		res, err = s.makeHelpMessage(parsed)
		if err != nil {
			s.logger.WithError(err).Error("failed to prepare response to help command")
			return
		}

		message, err = json.Marshal(res)
		if err != nil {
			s.logger.WithError(err).Error("failed to marshal response to command")
			return
		}

		err = s.sendSlackResponse(parsed.ResponseURL, message)
		if err != nil {
			s.logger.WithError(err).Error("failed to reply to command")
			return
		}
		return
	}

	action, err := s.determineTriggeredAction(parsed)
	if err != nil {
		if err == errCommandUndetermined {
			s.logger.WithError(err).WithField("parsed", parsed).Error()
			return
		}
		s.logger.WithError(err).Error("unknown error occurred")
		return
	}

	res, err = action(parsed)
	if err != nil {
		s.logger.WithError(err).Error("an error occurred while prepping a response to the command")
		return
	}

	message, err = json.Marshal(res)
	if err != nil {
		s.logger.WithError(err).Error("failed to marshal response to command")
		return
	}

	err = s.sendSlackResponse(parsed.ResponseURL, message)
	if err != nil {
		s.logger.WithError(err).Error("failed to reply to command")
		return
	}
}

func (s *service) determineTriggeredAction(parsed ParsedCommand) (Action, error) {

	for _, com := range s.flat {

		if com.Strict && strSliceContainsString(com.Triggers, parsed.Text) {
			return com.Action, nil
		}
		if !com.Strict && strings.HasPrefix(parsed.Text, com.Prefix) {
			return com.Action, nil
		}
	}

	return nil, errCommandUndetermined

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
