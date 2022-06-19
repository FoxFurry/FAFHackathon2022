package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
)

const (
	issuerKey = "iss"
	uuidKey   = "uuid"
	expKey    = "exp"
	subKey    = "sub"
)

type JWTProvider interface {
	ValidateToken(token string) (string, error)
}

type jwtprovider struct {
	issuer string
	key    []byte
}

func NewJWT(issuer string, key []byte) JWTProvider {
	return &jwtprovider{
		issuer: issuer,
		key:    key,
	}
}

func (j *jwtprovider) ValidateToken(token string) (string, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}

		return j.key, nil
	})
	if err != nil {
		return "", err
	}

	var sub string
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		//if claims[issuerKey] != j.issuer {
		//	return "", errors.New("issuers mismatch")
		//}
		sub, ok = claims[subKey].(string)
		if !ok {
			return "", errors.New("missing sub claim")
		}
	}

	return sub, nil
}
