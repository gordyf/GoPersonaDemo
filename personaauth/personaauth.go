package personaauth

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var hashKey = []byte("12345678901234567890123456789012")
var blockKey = []byte("12345678901234567890123456789012")
var s = securecookie.New(hashKey, blockKey)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:  "auth",
		Value: "",
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
}

func getAudience(c appengine.Context) string {
	if appengine.IsDevAppServer() {
		return "http://localhost:8080/"
	} else {
		hostname := appengine.DefaultVersionHostname(c)
		return fmt.Sprintf("http://%v", hostname)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	a := r.FormValue("assertion")

	client := urlfetch.Client(c)
	values := make(url.Values)
	values.Set("audience", getAudience(c))
	values.Set("assertion", a)
	rv, err := client.PostForm("https://verifier.login.persona.org/verify", values)
	defer rv.Body.Close()

	if err != nil {
		c.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, _ := ioutil.ReadAll(rv.Body)

	type Verify struct {
		Audience, Issuer, Email, Status string
		Expires                         int64
	}
	var v Verify
	err = json.Unmarshal(body, &v)
	if err != nil {
		c.Errorf("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if v.Status == "okay" {
		// logged in!
		value := map[string]string{
			"userid": v.Email,
		}
		if encoded, err := s.Encode("auth", value); err == nil {
			msInS := int64(time.Second / time.Millisecond)
			cookie := http.Cookie{
				Name:    "auth",
				Value:   encoded,
				Path:    "/",
				Expires: time.Unix(v.Expires/msInS, 0),
			}
			http.SetCookie(w, &cookie)
		} else {
			c.Errorf("%v", err)
		}
	} else {
		c.Errorf("Failed login state: %v", v.Status)
	}
}

func GetLoggedInUser(r *http.Request) string {
	if cookie, err := r.Cookie("auth"); err == nil {
		value := make(map[string]string)
		if err = s.Decode("auth", cookie.Value, &value); err == nil {
			return value["userid"]
		}
	}
	return ""
}
