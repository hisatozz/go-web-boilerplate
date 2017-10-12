package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"gopkg.in/unrolled/secure.v1"
	"gopkg.in/yaml.v2"
)

/* for serverConfig.yml */
type AplicationConfig struct {
	OauthClientId     string `yaml:"oauthClientId"`
	OauthClientSecret string `yaml:"oauthClientSecret"`
	RedirectUrl       string `yaml:"redirectUrl"`
	Key               string `yaml:"key"`
	TokenTTL          int    `yaml:"tokenTTL"`
}

func readConfig(filename string) (*AplicationConfig, error) {
	fileStr, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config AplicationConfig
	err = yaml.Unmarshal(fileStr, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

const staticURL = "http://www.local.test"

func main() {

	log.SetLevel(log.DebugLevel)

	secureMd := secure.New(secure.Options{
		AllowedHosts:          []string{"www.local.test", "www2.local.test", "api.local.test"},
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self' http://api.local.test https://api.local.test http://fonts.googleapis.com http://fonts.gstatic.com",
		IsDevelopment:         false,
	})

	coMd := cors.New(cors.Options{
		AllowedOrigins:   []string{staticURL},
		AllowedMethods:   []string{"GET", "PUT", "POST"},
		AllowCredentials: false,
		Debug:            true,
	})

	sv := NewWebApp()

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(negroni.NewRecovery())
	n.Use(negroni.HandlerFunc(secureMd.HandlerFuncWithNext))
	n.Use(coMd)
	n.Use(negroni.NewStatic(http.Dir("public_html")))

	router := httprouter.New()
	router.GET("/hello/:name", sv.helloHandle())
	router.GET("/login", sv.loginHandle())
	router.GET("/oauth-redirect", oauthRedirectHandler)
	router.POST("/api/token-exchange", sv.tokenExchangeAPI())
	router.GET("/api/hello", HelloAPI)
	router.PUT("/api/hello", HelloAPI)
	router.POST("/api/private/hello", sv.somePrivateApi())

	n.UseHandler(router)

	log.Fatal(http.ListenAndServe(":80", n))
}
