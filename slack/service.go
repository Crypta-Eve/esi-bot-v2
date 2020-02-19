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
	HandleSlashCommand(context.Context, SlashCommand)
}

type service struct {
	logger   *logrus.Logger
	commands []Category
	flat     []Command
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
	res       Msg
	message   []byte
	err       error
	layoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
)

func (s *service) HandleSlashCommand(ctx context.Context, command SlashCommand) {
	time.Sleep(time.Millisecond * 250)

	// Check to see if this is a call for help
	if command.Text == "" || command.Text == "help" {

		res, err = s.makeHelpMessage(command)
		if err != nil {
			s.logger.WithError(err).Error("failed to prepare response to help command")
			return
		}

		message, err = json.Marshal(res)
		if err != nil {
			s.logger.WithError(err).Error("failed to marshal response to command")
			return
		}

		err = s.sendSlackResponse(command.ResponseURL, message)
		if err != nil {
			s.logger.WithError(err).Error("failed to reply to command")
			return
		}
		return
	}

	action, err := s.determineTriggeredAction(command)
	if err != nil {
		if _, ok := knownErrs[err]; ok {
			s.logger.WithError(err).WithField("command", command).Error()
			return
		}
		s.logger.WithError(err).Error("unknown error occurred")
		return
	}

	res, err := action(command)
	if err != nil {
		s.logger.WithError(err).Error("an error occurred while prepping a response to the command")
		return
	}

	message, err = json.Marshal(res)
	if err != nil {
		s.logger.WithError(err).Error("failed to marshal response to command")
		return
	}

	err = s.sendSlackResponse(command.ResponseURL, message)
	if err != nil {
		s.logger.WithError(err).Error("failed to reply to command")
		return
	}
}

func (s *service) determineTriggeredAction(command SlashCommand) (Action, error) {

	var triggered *Command

	for _, com := range s.flat {

		if com.Strict && strSliceContainsString(com.Triggers, command.Text) {
			triggered = &com
			break
		}
		if !com.Strict && strings.HasPrefix(command.Text, com.Prefix) {
			triggered = &com
			break
		}
	}
	if triggered == nil {
		return nil, errCommandUndetermined
	}

	if len(triggered.Args) == 0 || len(command.Args) == 0 {
		return triggered.Action, nil
	}

	for comArgKey, comArgVal := range command.Args {
		if _, ok := triggered.Args[comArgKey]; !ok {
			return nil, errCommandWithInvalidArgs
		}
		isValidValue := false
		for _, trigVal := range triggered.Args[comArgKey] {
			if trigVal == comArgVal {
				isValidValue = true
				break
			}
		}
		if !isValidValue {
			return nil, errCommandWithInvalidArgValue
		}
	}

	return triggered.Action, nil

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
