package sessions

import (
	"encoding/gob"
	"encoding/hex"
	"errors"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/auth"
)

func SetupSessionMiddleware(authKey string, encryptionKey string) (gin.HandlerFunc, error) {
	var profile *auth.UserProfile
	gob.Register(profile)

	sessionAuthKey, err := hex.DecodeString(authKey)
	if err != nil {
		log.Fatalf("Could not decode auth key")
	}
	sessionEncryptionKey, err := hex.DecodeString(encryptionKey)
	if err != nil {
		log.Fatalf("Could not decode encryption key")
	}
	store := cookie.NewStore(sessionAuthKey, sessionEncryptionKey)
	middleware := sessions.Sessions("auth-session", store)
	return middleware, nil
}

var ErrProfileNotValid = errors.New("profile can not be parsed")

func GetProfile(c *gin.Context) (*auth.UserProfile, error) {
	p, ok := sessions.Default(c).Get("profile").(*auth.UserProfile)
	if !ok {
		return nil, ErrProfileNotValid
	}
	return p, nil
}
