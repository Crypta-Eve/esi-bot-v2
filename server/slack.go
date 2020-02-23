package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nlopes/slack/slackevents"

	goslack "github.com/nlopes/slack"
	"github.com/pkg/errors"
)

func (s *Server) handlePostSlack(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	err := verifySlackReqeust(r, s.Config.SlackSigningSecret)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	body := buf.Bytes()
	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	if event.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal(body, &r)
		if err != nil {
			s.WriteError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		s.WriteSuccess(ctx, w, r, http.StatusOK)
		return
	}

	go func(event slackevents.EventsAPIEvent) {
		// Delay for dramatic effect
		time.Sleep(time.Millisecond * 500)
		switch e := event.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			s.Slack.HandleMessageEvent(ctx, e)
		}
	}(event)

	s.WriteSuccess(ctx, w, nil, http.StatusOK)

}

func verifySlackReqeust(req *http.Request, secret string) error {
	verifier, err := goslack.NewSecretsVerifier(req.Header, secret)
	if err != nil {
		return errors.Wrap(err, "failed to create secrets verifier")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read the body from the request")
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	_, err = verifier.Write(body)
	if err != nil {
		return errors.Wrap(err, "failed to write body to the verifier")
	}

	err = verifier.Ensure()
	if err != nil {
		return errors.Wrap(err, "failed to ensure the verifier")
	}

	return nil
}
