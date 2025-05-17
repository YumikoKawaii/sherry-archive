package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Resolver interface {
	GenerateToken(params TokenParameters) (string, error)
	ResolveToken(token string) (*Claims, error)
}

type Claims struct {
	UserId int64    `json:"userId"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type TokenParameters struct {
	UserId int64
	Roles  []string
}

type resolverImpl struct {
	key      string
	ttlInSec int64
}

func NewResolver(key string, ttlInSec int64) Resolver {
	return &resolverImpl{
		key:      key,
		ttlInSec: ttlInSec,
	}
}

func (r *resolverImpl) GenerateToken(params TokenParameters) (string, error) {
	expirationTime := time.Now().Add(time.Second * time.Duration(r.ttlInSec))
	claims := &Claims{
		UserId: params.UserId,
		Roles:  params.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(r.key)
}

func (r *resolverImpl) ResolveToken(token string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return r.key, nil
	})
	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, err
}
