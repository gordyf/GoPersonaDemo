package hello

import (
	"appengine"
	"appengine/datastore"
	"html/template"
	"net/http"
	"personaauth"
	"time"
)

type TemplateData struct {
	Greetings []Greeting
	Userid    string
}

type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

var guestbookTemplate = template.Must(template.ParseFiles("templates/root.html"))

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/login", personaauth.LoginHandler)
	http.HandleFunc("/logout", personaauth.LogoutHandler)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := datastore.NewQuery("Greeting").Order("-Date").Limit(10)
	greetings := make([]Greeting, 0, 10)
	if _, err := q.GetAll(c, &greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u := personaauth.GetLoggedInUser(r)
	td := TemplateData{greetings, u}
	if err := guestbookTemplate.Execute(w, td); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Greeting{
		Content: r.FormValue("content"),
		Date:    time.Now(),
		Author:  personaauth.GetLoggedInUser(r),
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Greeting", nil), &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
