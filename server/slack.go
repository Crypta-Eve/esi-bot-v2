package server

import (
	"net/http"

	"github.com/eveisesi/eb2/slack"
	"github.com/mitchellh/mapstructure"
)

func (s *Server) handlePostSlack(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	// data, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	s.WriteError(ctx, w, err, http.StatusInternalServerError)
	// 	return
	// }

	err := r.ParseForm()
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}
	var halfparse = map[string]string{}
	for i, r := range r.Form {
		halfparse[i] = r[0]
	}

	var parsed slack.ParsedCommand
	err = mapstructure.Decode(halfparse, &parsed)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	s.WriteSuccess(ctx, w, nil, http.StatusOK)

	go s.Slack.HandleSlashCommand(ctx, parsed)

}
