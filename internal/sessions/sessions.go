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

	session_auth_key, err := hex.DecodeString(authKey)
	if err != nil {
		log.Fatalf("Could not decode auth key")
	}
	session_encryption_key, err := hex.DecodeString(encryptionKey)
	if err != nil {
		log.Fatalf("Could not decode encryption key")
	}
	store := cookie.NewStore(session_auth_key, session_encryption_key)
	middleware := sessions.Sessions("auth-session", store)
	return middleware, nil
}

var (
	ProfileNotValid = errors.New("profile can not be parsed")
)

func GetProfile(c *gin.Context) (*auth.UserProfile, error) {
	p, ok := sessions.Default(c).Get("profile").(*auth.UserProfile)
	if !ok {
		return nil, ProfileNotValid
	}
	return p, nil
}
