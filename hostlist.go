package hostlist

import (
	"appengine"
	"appengine/user"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	demandLogin = true
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/addHost", addHost)
	http.HandleFunc("/delete", delete)
}

func root(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if demandLogin && getUser(ctx, w, r) == nil {
		return
	}
	result, err := NewDbHandle(ctx).Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := reportTemplate.Execute(w, result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Should use an xsrf token here.  Meh.
func addHost(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		http.Error(w, "No user!", http.StatusInternalServerError)
		return
	}
	g := HostRecord{
		Content: r.FormValue("content"),
		Date:    time.Now(),
		Author:  u.String(),
	}
	err := NewDbHandle(c).Write(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("idToDelete"), 10, 64)
	log.Printf("Id to delete = %d", id)
	err = NewDbHandle(appengine.NewContext(r)).Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

var reportTemplate = template.Must(template.New("whatevs").Parse(reportTemplateHTML))

const reportTemplateHTML = `
<html>
  <head>
    <title>Volley Hosts</title>
  </head>
  <body>
    <form action="/addHost" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Add namespace.root"></div>
    </form>
    {{range .}}
      {{with .Author}}
        <b>{{.}}</b>
      {{else}}
        anonymous
      {{end}}
      :  v23.namespace.root={{.Content}}
      <form action="/delete" method="post">
        <input type="hidden" name="idToDelete" value="{{.Id}}">
        <input type="submit" value="delete it">
      </form>
    {{end}}
  </body>
</html>
`

func getUser(
	ctx appengine.Context,
	w http.ResponseWriter,
	req *http.Request) *user.User {
	u := user.Current(ctx)
	if u != nil {
		return u
	}
	url, err := user.LoginURL(ctx, req.URL.String())
	if err == nil {
		// Redirect to login.
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
	} else {
		// Show error page.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return nil
}
