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
	chatty      = false
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/addHost", addHost)
	http.HandleFunc("/deleteHost", deleteHost)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if demandLogin && getUser(c, w, r) == nil {
		return
	}
	result, err := NewDbHandle(c).Read()
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
	u := getUser(c, w, r)
	if u == nil {
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

func deleteHost(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := getUser(c, w, r)
	if u == nil {
		return
	}
	id, err := strconv.ParseInt(r.FormValue("idToDelete"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if chatty {
		log.Printf("Id to delete = %d", id)
	}
	err = NewDbHandle(c).Delete(id)
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
      <textarea name="content" rows="2" cols="60"></textarea>
      <input type="submit" value="Add another">
    </form>
    <table>
    {{range .}}
      <tr>
        <td>
          v23.namespace.root={{.Content}} 
        </td>
        <td>
          {{.TimeNice}}
        </td>
        <td>
          {{with .Author}} <b>{{.}}</b>  {{else}} anonymous {{end}}
        </td>
        <td>
          <form action="/deleteHost" method="post">
          <input type="hidden" name="idToDelete" value="{{.Id}}">
          <input type="submit" value="delete this">
          </form>
        </td>
      </tr>
    {{end}}
    </table>
  </body>
</html>
`

func getUser(
	ctx appengine.Context,
	w http.ResponseWriter,
	req *http.Request) *user.User {
	u := user.Current(ctx)
	if u != nil {
		if isAuthorized(u) {
			return u
		} else {
			http.Error(w, u.String()+" unauthorized.", http.StatusUnauthorized)
			return nil
		}
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

var authorizedUsers = []string{"test@example.com", "jeff.regan"}

func isAuthorized(u *user.User) bool {
	author := u.String()
	for _, item := range authorizedUsers {
		if author == item {
			return true
		}
	}
	return false
}
