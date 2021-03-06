package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// handle request on /login
func (sv *WebApp) loginHandle() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		cf := sv.oauth2Config

		state := sv.generateStateToken()
		urlstr := cf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		log.Debug("Visit the URL for the auth dialog: %v", urlstr)

		w.Header().Add("Location", urlstr)
		w.WriteHeader(302)
	}
}

// handle request on /oauth-redirect
// oauthRedirectHandler handle redirect from github
func oauthRedirectHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := r.FormValue("code")
	state := r.FormValue("state")

	log.Debugf("code: %s\tstate:%s", code, state)

	vl := &url.Values{}
	vl.Set("code", code)
	vl.Add("state", state)

	urlobj, _ := url.Parse(staticURL + "/token-exchange.html")
	urlobj.RawQuery = vl.Encode()

	log.Debugf("redirect to %s", urlobj)

	w.Header().Add("Location", urlobj.String())
	w.WriteHeader(302)
	return
}

// handle request on /api/token-exchange
// The two arguments "code" and "state" are passed from github API at oauthRedirectHandler.
func (sv *WebApp) tokenExchangeAPI() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		err := func() error {
			var state, code, loginname, token string
			var jsData *jason.Object
			var githubToken *oauth2.Token
			var client *http.Client
			var resp *http.Response
			var err error
			var user *jason.Object
			var uid UID

			// read json
			jsData, err = jason.NewObjectFromReader(io.LimitReader(r.Body, 1024))
			if err != nil {
				return err
			}

			state, err = jsData.GetString("state")
			log.Debugf("tokenExchangeAPI() state: %s", state)

			code, err = jsData.GetString("code")
			log.Debugf("code: %s", code)

			if err != nil {
				return err
			}

			// 1. check state token

			err = sv.checkStateToken(state)
			if err != nil {
				return err
			}

			// 2. get github user data

			githubToken, err = sv.oauth2Config.Exchange(r.Context(), code)
			if err != nil {
				log.Infof("tokenExchangeAPI(): Error in Exchange. %v", err)
				return err
			}
			log.Debug(githubToken)
			client = sv.oauth2Config.Client(r.Context(), githubToken)
			resp, err = client.Get("https://api.github.com/user")
			if err != nil {
				log.Infof("tokenExchangeAPI(): Error in get /user. %v", err)
				return err
			}

			// make response
			user, _ = jason.NewObjectFromReader(resp.Body)
			loginname, _ = user.GetString("login")

			log.Debug(loginname)

			uid = getUserIDByGithubName(loginname)
			token, createdTime := sv.generateToken(uid)
			expire := createdTime.Add(time.Duration(sv.appConfig.TokenTTL) * time.Minute)
			log.Debugf("created: %s, expire: %s", createdTime.Format(time.RFC3339), expire.Format(time.RFC3339))

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"token": "%s", "githubName":"%s", "expire": "%s"}`, token, loginname, expire.Format("2006-01-02T15:04:05-0700"))

			return err
		}()
		if err != nil {
			w.WriteHeader(422)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"type":"Authentication Error", "message":"%s"}`, err)

		}

	}

}

/*
generateStateToken() generate state for github OAuth.
The first 8bytes are current UNIX Time in sec. Next 16 bytes are random number and the last 32 bytes are MAC of time and random number.
MAC is generated by HMAC-SHA256
*/
func (sv *WebApp) generateStateToken() string {
	// make 64byte length randomnumber
	buf := new(bytes.Buffer)
	buf.Grow(8 + 16 + 32)

	created := time.Now().Unix()
	binary.Write(buf, binary.LittleEndian, created)

	io.CopyN(buf, rand.Reader, 16)

	mac := hmac.New(sha256.New, sv.key)
	mac.Write(buf.Bytes())
	hm := mac.Sum(nil)

	log.Debugf("generateStateToken() token: %x\thmac: %x", buf.Bytes(), hm)
	buf.Write(hm)
	log.Debugf("generateStateToken() state: %x", buf.Bytes())

	return fmt.Sprintf("%x", buf.Bytes())
}

// check state token by hmac and expiration time
func (sv *WebApp) checkStateToken(state string) error {
	// 1. Length check
	if len(state) != ((8 + 16 + 32) * 2) {
		log.Infof("checkStateToken() len(state) = %d", len(state))
		return fmt.Errorf("length of state param is invalid.len:%d", len(state))
	}

	return sv.checkToken(state)
}

/*
 * generateToken() generate token for private API.
 * First 8bytes are timestamp.
 * Next sizeof(UID) is userid.
 * Last 32 bytes are hmac.
 */
func (sv *WebApp) generateToken(uid UID) (string, time.Time) {
	buf := new(bytes.Buffer)

	created := time.Now().Unix()
	binary.Write(buf, binary.LittleEndian, created)
	binary.Write(buf, binary.LittleEndian, uid)

	mac := hmac.New(sha256.New, sv.key)
	mac.Write(buf.Bytes())
	hm := mac.Sum(nil)

	buf.Write(hm)
	log.Debugf("generateToken() token: %x", buf.Bytes())

	return fmt.Sprintf("%x", buf.Bytes()), time.Unix(created, 0)
}

// check token by hmac and expiration time
func (sv *WebApp) checkToken(tokenStr string) error {

	// 1. decode hex
	b := make([]byte, hex.DecodedLen(len(tokenStr)))
	_, err := hex.Decode(b, []byte(tokenStr))
	if err != nil {
		return err
	}

	// 2. Expiration check
	var createdTime time.Time
	var createdInt64 int64
	err = binary.Read(bytes.NewReader(b[:8]), binary.LittleEndian, &createdInt64)
	if err != nil {
		return err
	}

	createdTime = time.Unix(createdInt64, 0)
	now := time.Now()

	dur := now.Sub(createdTime)
	durInt := int(dur.Minutes())

	if durInt > sv.appConfig.TokenTTL {
		return fmt.Errorf("login expired. %s", tokenStr)
	}

	// 3. hmac check
	size := len(b)
	payload := b[8 : size-32]
	data := b[:size-32]
	mac := b[size-32:]

	log.Debugf("token was created at %s, payload: %x, hmac is %x", createdTime, payload, mac)

	codec := hmac.New(sha256.New, sv.key)
	codec.Write(data)
	expectedMac := codec.Sum(nil)

	ok := hmac.Equal(mac, expectedMac)

	if ok {
		log.Info("checkToken succeed.")
	} else {
		log.Infof("checkToken() hmac check faild. token data:%x\t expectedMac:%x", data, expectedMac)
		return fmt.Errorf("checkToken() hmac check faild. token data:%x\t expectedMac:%x", data, expectedMac)
	}
	return nil
}
