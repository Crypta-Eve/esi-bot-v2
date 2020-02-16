package server

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

func (s *Server) handlePostSlack(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	parsed, err := url.ParseQuery(string(data))
	if err != nil {
		s.WriteError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	go s.Slack.HandleSlashCommand(ctx, parsed)

	// response := struct {
	// 	Text string `json:"text"`
	// }{
	// 	Text: "Thinking....Gimme a Sec",
	// }

	s.WriteSuccess(ctx, w, nil, http.StatusOK)

}
