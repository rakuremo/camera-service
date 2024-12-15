package middleware

import (
	"net/http"

	"github.com/deepch/RTSPtoWeb/config"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

type Message struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

var store *sessions.CookieStore
var log = logrus.New()

func init() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err.Error())
	}

	store = sessions.NewCookieStore([]byte(cfg.AuthConfig.SessionKey))
}

func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request

		session, err := store.Get(r, "auth")
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, Message{Status: 500, Payload: "cannot load authorized data from request"})
			c.Abort()
			return
		}

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetStore() *sessions.CookieStore { 
	return store
}
