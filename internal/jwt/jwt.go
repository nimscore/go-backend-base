package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenInvalid = errors.New("token invalid")
var ErrTokenExpired = errors.New("token expired")
var ErrTokenKindInvalid = errors.New("token kind invalid")
var ErrTokenSchemaMalformed = errors.New("token schema malformed")

const KIND_ACCESS = "access"
const KIND_REFRESH = "refresh"

const ACCESS_TOKEN_EXPIRATION = 15 * time.Minute
const REFRESH_TOKEN_EXPIRATION = 7 * 24 * time.Hour

type JWT struct {
	secret []byte
}

func NewJWT(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

func (this *JWT) GenerateAccessToken(sessionID string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"session_id": sessionID,
			"expiration": time.Now().Add(ACCESS_TOKEN_EXPIRATION).Unix(),
			"kind":       KIND_ACCESS,
		},
	)
	return token.SignedString(this.secret)
}

func (this *JWT) GenerateRefreshToken(sessionID string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"session_id": sessionID,
			"expiration": time.Now().Add(REFRESH_TOKEN_EXPIRATION).Unix(),
			"kind":       KIND_REFRESH,
		},
	)
	return token.SignedString(this.secret)
}

func (this *JWT) ParseAccessToken(token string) (string, error) {
	claims, err := this.parseToken(token)
	if err != nil {
		return "", err
	}

	kind, ok := claims["kind"].(string)
	if !ok || kind != "access" {
		return "", ErrTokenKindInvalid
	}

	id, ok := claims["session_id"].(string)
	if !ok {
		return "", ErrTokenSchemaMalformed
	}

	return id, nil
}

func (this *JWT) ParseRefreshToken(token string) (string, error) {
	claims, err := this.parseToken(token)
	if err != nil {
		return "", err
	}

	kind, ok := claims["kind"].(string)
	if !ok || kind != "refresh" {
		return "", ErrTokenKindInvalid
	}

	id, ok := claims["session_id"].(string)
	if !ok {
		return "", ErrTokenSchemaMalformed
	}

	return id, nil
}

func (this *JWT) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return this.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenSchemaMalformed
	}

	expiration, ok := claims["expiration"].(float64)
	if !ok {
		return nil, ErrTokenSchemaMalformed
	}

	if int64(expiration) < time.Now().Unix() {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
