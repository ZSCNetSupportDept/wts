package hutil

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"zsxyww.com/wts/model/sqlc"
)

type WtsJWT struct {
	OpenID   string         `json:"openid"`
	Sid      string         `json:"sid"`
	Username string         `json:"username"`
	Avatar   string         `json:"avatar"`
	Access   sqlc.WtsAccess `json:"access"`
	Name     string         `json:"name"`
	jwt.RegisteredClaims
}

var NewWtsJWT func(openID, sid string, access sqlc.WtsAccess, username string, avatar string, name string, expire int) (string, error)

func InitJWTKey(key string) {
	NewWtsJWT = func(openID, sid string, access sqlc.WtsAccess, username string, avatar string, name string, expire int) (string, error) {

		t := &WtsJWT{
			OpenID:   openID,
			Sid:      sid,
			Access:   access,
			Username: username,
			Avatar:   avatar,
			Name:     name,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		a := jwt.NewWithClaims(jwt.SigningMethodHS256, t)

		token, err := a.SignedString([]byte(key))
		if err != nil {
			return "", err
		}
		return token, nil
	}

}
