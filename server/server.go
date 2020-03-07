package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/eveisesi/eb2/slack"
	"github.com/eveisesi/eb2/token"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

type Server struct {
	server  *http.Server
	Config  *eb2.Config
	Logger  *logrus.Logger
	Slack   slack.Service
	Token   token.Service
	goslack *nslack.Client
}

func NewServer(config *eb2.Config, logger *logrus.Logger, slack slack.Service, token token.Service) *Server {
	return &Server{
		Config:  config,
		Logger:  logger,
		Slack:   slack,
		Token:   token,
		goslack: nslack.New(config.SlackAPIToken),
	}
}

func (s *Server) Run() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Config.ApiPort),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      s.BuildRouter(),
	}
	s.Logger.WithField("port", s.Config.ApiPort).Infof("starting server on port: %d", s.Config.ApiPort)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) BuildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(Cors)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(NewStructuredLogger(s.Logger))

	r.Get("/slack/invite", s.handleGetSlackInvite)
	r.Post("/slack/invite", s.handlePostSlackInvite)
	r.Group(func(r chi.Router) {
		r.Use(s.CheckJWT)
		r.Post("/slack/invite/send", s.handlePostSlackInviteSend)
	})
	r.Post("/slack", s.handlePostSlack)

	return r

}

func (s *Server) WriteSuccess(ctx context.Context, w http.ResponseWriter, data interface{}, status int) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}

}

func (s *Server) WriteError(ctx context.Context, w http.ResponseWriter, data error, status int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		msg := struct {
			Message string `json:"message"`
		}{
			Message: data.Error(),
		}

		_ = json.NewEncoder(w).Encode(msg)

	}

}
