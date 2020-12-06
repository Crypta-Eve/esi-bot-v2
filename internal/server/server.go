package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/eveisesi/eb2/internal/slack"
	"github.com/eveisesi/eb2/internal/token"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	nslack "github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

type server struct {
	server  *http.Server
	config  *eb2.Config
	logger  *logrus.Logger
	slack   slack.Service
	token   token.Service
	goslack *nslack.Client
}

func NewServer(config *eb2.Config, logger *logrus.Logger, slack slack.Service, token token.Service) *server {
	return &server{
		config:  config,
		logger:  logger,
		slack:   slack,
		token:   token,
		goslack: nslack.New(config.SlackAPIToken),
	}
}

func (s *server) Run() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.ApiPort),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      s.BuildRouter(),
	}
	s.logger.WithField("port", s.config.ApiPort).Infof("starting server on port: %d", s.config.ApiPort)
	return s.server.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *server) BuildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(Cors)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(NewStructuredLogger(s.logger))

	r.Get("/slack/invite", s.handleGetSlackInvite)
	r.Post("/slack/invite", s.handlePostSlackInvite)
	r.Group(func(r chi.Router) {
		r.Use(s.CheckJWT)
		r.Post("/slack/invite/send", s.handlePostSlackInviteSend)
	})
	r.Post("/slack", s.handlePostSlack)

	return r

}

func (s *server) writeSuccess(ctx context.Context, w http.ResponseWriter, data interface{}, status int) {

	w.WriteHeader(status)

	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}

}

func (s *server) writeError(ctx context.Context, w http.ResponseWriter, data error, status int) {

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
