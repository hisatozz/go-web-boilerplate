package main

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/antonholmquist/jason"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type WebApp struct {
	appConfig    *AplicationConfig
	oauth2Config *oauth2.Config
	key          []byte
}

func NewWebApp() *WebApp {
	sv := &WebApp{}
	sc, err := readConfig("./appConfig.yaml")
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("%+v", sc)

	o2c := &oauth2.Config{
		ClientID:     sc.OauthClientId,
		ClientSecret: sc.OauthClientSecret,
		RedirectURL:  sc.RedirectUrl,
		Scopes:       []string{"openid", "email"},
		Endpoint:     github.Endpoint,
	}

	sv.appConfig = sc
	sv.oauth2Config = o2c
	sv.key = []byte(sv.appConfig.Key)

	return sv
}

type UID int32

// handle request on /hello
func (sv *WebApp) helloHandle() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	}
}

// handle request on /api/hello
func HelloAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"hello\": \"world\"}"))
	log.Debug("HelloAPI called.")
}

func getUserIDByGithubName(uid string) UID {
	return 1
}

// handle "/api/private/hello"
// It takes one argument: token string
func (sv *WebApp) somePrivateApi() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := func() error {
			// Parse JSON
			obj, err := jason.NewObjectFromReader(io.LimitReader(r.Body, 1024))
			if err != nil {
				return err
			}

			token, err := obj.GetString("token")
			if err != nil {
				return err
			}

			// Check Token
			err = sv.checkToken(token)
			if err != nil {
				return err
			}

			// Return Value
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"hello": "private world"}`)
			return nil
		}()

		if err != nil {
			w.WriteHeader(422)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"type":"Authentication Error", "message":"%s"}`, err)
			return
		}
	}
}
