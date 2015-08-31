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
	chatty = false
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/addHost", addHost)
	http.HandleFunc("/deleteHost", deleteHost)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Want to show hostnames to anyone, but only allow authorized users
	// to see extra data, and get controls to add and delete.
	result, err := NewDbHandle(c).Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u := user.Current(c)
	data := struct {
		IsAuthorized bool
		Records      []HostRecord
	}{
		u != nil && isAuthorized(u),
		result,
	}
	if err := reportTemplate.Execute(w, data); err != nil {
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
    {{if $.IsAuthorized}}
      <form action="/addHost" method="post">
        <input type="text" name="content"></textarea>
        <input type="submit" value="Add another">
      </form>
    {{end}}
    <table>
      {{range .Records}}
      <tr>
        <td>
          v23.namespace.root={{.Content}} 
        </td>
        {{if $.IsAuthorized}}
          <td>
            {{.TimeNice}}
          </td>
          <td>
            {{with .Author}} <b>{{.}}</b> {{else}} anonymous {{end}}
          </td>
          <td>
            <form action="/deleteHost" method="post">
            <input type="hidden" name="idToDelete" value="{{.Id}}">
            <input type="submit" value="delete this">
            </form>
          </td>
        {{end}}
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
