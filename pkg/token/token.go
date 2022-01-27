package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// FetchTokenLifeTime fetches token life time of the token
func FetchTokenLifeTime(tokenString string) (uint64, error) {
	var tokenLifeTime uint64

	token, err := parseToken(tokenString)
	if err != nil {
		return tokenLifeTime, errors.New("error parsing the token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if err := claims.Valid(); err != nil {
			return tokenLifeTime, err
		}
		currentTime := time.Now().Unix()
		var expiryTime interface{}
		if expiryTime, ok = claims["exp"]; !ok {
			return tokenLifeTime, errors.New("unable to find expiry time of token")
		}
		tokenLifeTime = uint64(expiryTime.(float64)) - uint64(currentTime)
		return tokenLifeTime, nil
	}
	return tokenLifeTime, errors.New("unable to fetch token claims")
}

// parseToken parses token string to jwt token
func parseToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, errors.New("empty token string")
	}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	return token, err
}
