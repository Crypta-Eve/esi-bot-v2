package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"

	"github.com/eveisesi/eb2/tools"
	"github.com/nlopes/slack/slackevents"
	"github.com/patrickmn/go-cache"

	goslack "github.com/nlopes/slack"
	"github.com/pkg/errors"
)

func (s *Server) handlePostSlack(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	err := verifySlackReqeust(r, s.Config.SlackSigningSecret)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusBadRequest)
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
		switch e := event.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			s.Slack.HandleMessageEvent(ctx, e)
		}
	}(event)

	s.WriteSuccess(ctx, w, nil, http.StatusOK)

}

var (
	stateMap = cache.New(time.Minute*5, time.Minute*5)
)

type (
	SlackInvite struct {
		State string `json:"state"`
		Code  string `json:"code"`
	}

	Token struct {
		Token string `json:"token"`
	}
)

func (s SlackInvite) IsValid() bool {
	return (s.State != "" && s.Code != "")
}

func (s *Server) handleGetSlackInvite(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	state := tools.RandomString(16)
	stateMap.Set(state, true, 0)

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("redirect_uri", s.Config.EveCallback)
	query.Set("client_id", s.Config.EveClientID)
	query.Set("state", state)

	uri := url.URL{
		Scheme:   "https",
		Host:     "login.eveonline.com",
		Path:     "/v2/oauth/authorize",
		RawQuery: query.Encode(),
	}

	s.WriteSuccess(ctx, w, struct {
		Url string `json:"url"`
	}{
		Url: uri.String(),
	}, http.StatusOK)

}

func (s *Server) handlePostSlackInvite(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var body SlackInvite
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusBadRequest)
		return
	}

	if !body.IsValid() {
		s.WriteError(ctx, w, errors.New("invalid body received. Please provide both code and state"), http.StatusBadRequest)
		return
	}

	_, found := stateMap.Get(body.State)
	if !found {
		s.WriteError(ctx, w, errors.New("invalid state received"), http.StatusBadRequest)
		return
	}

	stateMap.Delete(body.State)

	uri := url.Values{}
	uri.Set("grant_type", "authorization_code")
	uri.Set("code", body.Code)

	req, err := http.NewRequest(http.MethodPost, "https://login.eveonline.com/v2/oauth/token", bytes.NewBuffer([]byte(uri.Encode())))
	if err != nil {
		s.WriteError(ctx, w, errors.Wrap(err, "unable to configure post request to ccp"), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "login.eveonline.com")
	req.Header.Set("User-Agent", "Tweetfleet Slack Inviter (david@onetwentyseven.dev || TF Slack @doubled)")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Config.EveClientID+":"+s.Config.EveClientSecret)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.WriteError(ctx, w, errors.Wrap(err, "failed to make post request to ccp"), http.StatusInternalServerError)
		return
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		s.WriteError(ctx, w, errors.Wrap(err, "unable to parse response from ccp"), http.StatusInternalServerError)
		return
	}

	s.WriteSuccess(ctx, w, data, http.StatusOK)

}

type SlackInviteSend struct {
	Email string `json:"email"`
}

type SlackInvitePayload struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
}

type SlackInviteResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (s *Server) handlePostSlackInviteSend(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var body SlackInviteSend
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		s.WriteError(ctx, w, errors.Wrap(err, "unable to decode request body"), http.StatusBadRequest)
		return
	}

	check := ctx.Value(tokenKey)
	if check == nil {
		s.WriteError(ctx, w, errors.New("token not found"), http.StatusInternalServerError)
		return
	}

	token := check.(*jwt.Token)

	endpoint := "https://slack.com/api/users.admin.invite"

	realName := token.Claims.(jwt.MapClaims)["name"].(string)

	uri := url.Values{}
	uri.Set("token", s.Config.SlackLegacyAPIToken)
	uri.Set("email", body.Email)
	uri.Set("real_name", realName)

	resp, err := http.PostForm(endpoint, uri)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var slackResp = &SlackInviteResponse{}
	err = json.NewDecoder(resp.Body).Decode(slackResp)
	if err != nil {
		s.WriteError(ctx, w, errors.Wrap(err, "unable to decode response from slack"), http.StatusInternalServerError)
		return
	}

	status := http.StatusOK

	switch slackResp.Ok {
	case true:
		msg := fmt.Sprintf("I've successfully invited %s (%s) to TF Slack.", realName, body.Email)
		channel, timestamp, err := s.goslack.PostMessage(s.Config.SlackModChannel, nslack.MsgOptionText(msg, false))
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"channel":   channel,
				"timestamp": timestamp,
				"message":   msg,
			}).Error("failed to post success message to mod chat.")
		}
	case false:
		status = http.StatusBadRequest
		msg := fmt.Sprintf("Uh Oh, I'm having issues inviting %s (%s) to TF Slack. Slack Response Dump: %s", realName, body.Email, slackResp.Error)
		channel, timestamp, err := s.goslack.PostMessage(s.Config.SlackModChannel, nslack.MsgOptionText(msg, false))
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"channel":   channel,
				"timestamp": timestamp,
				"message":   msg,
			}).Error("failed to post message to mod chat.")
		}
	}

	data, err := json.Marshal(slackResp)

	w.WriteHeader(status)
	w.Write(data)

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
