package token

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	"github.com/pkg/errors"
)

func (s *service) ParseAndValidateToken(token string) (*jwt.Token, error) {
	parser := new(jwt.Parser)
	parser.UseJSONNumber = true

	return parser.Parse(token, s.getKey)
}

func (s *service) getKey(token *jwt.Token) (interface{}, error) {
	var keys []byte = nil

	cachedKeys, found := s.caches["keys"].Get("keys")
	if !found {
		keys, err = s.fetchAndCacheJWKS()
		if err != nil {
			return nil, fmt.Errorf("keys missing from cache and unable to fetch fresh set of keys")
		}
	}

	if keys == nil && cachedKeys != nil {
		keys = cachedKeys.([]byte)
	}

	set, err := jwk.Parse(keys)
	if err != nil {
		return nil, err
	}

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("expecting JWT header to have string kid")
	}

	if key := set.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, fmt.Errorf("unable to find key")

}

func (s *service) fetchAndCacheJWKS() ([]byte, error) {

	uri, _ := url.Parse(jwksURL)

	resp, err := s.client.Get(uri.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch jwks")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode jwks body")

	}

	resp.Body.Close()

	s.caches["keys"].Set("keys", data, 0)

	return data, nil
}
