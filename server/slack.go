package server

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eveisesi/eb2/slack"
)

func (s *Server) handlePostSlack(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	query, err := url.ParseQuery(string(data))
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusBadRequest)
		return
	}

	parsed, err := slack.ParseSlashCommand(query)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusBadRequest)
		return
	}

	s.WriteSuccess(ctx, w, nil, http.StatusOK)

	go s.Slack.HandleSlashCommand(ctx, parsed)
}
