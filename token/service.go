package token

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/patrickmn/go-cache"
)

const jwksURL = "https://login.eveonline.com/oauth/jwks"

var (
	err error
)

type (
	Service interface {
		ParseAndValidateToken(token string) (*jwt.Token, error)
	}
	service struct {
		client *http.Client
		caches map[string]*cache.Cache
		config config
	}
	config struct {
		clientID     string
		clientSecret string
		userAgent    string
	}
)

func New(userAgent, clientID, clientSecret string) Service {

	http := &http.Client{
		Timeout: 30 * time.Second,
	}

	s := &service{
		client: http,
		caches: map[string]*cache.Cache{
			"keys": cache.New(24*time.Hour, 36*time.Hour),
		},
		config: config{
			clientID:     clientID,
			clientSecret: clientSecret,
			userAgent:    userAgent,
		},
	}

	_, err := s.fetchAndCacheJWKS()
	if err != nil {
		panic(err)
	}

	return s
}
