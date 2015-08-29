package hostlist

import (
	"appengine"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const (
	demandLogin = true
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/show", show)
	http.HandleFunc("/form", form)
	http.HandleFunc("/sign", sign)
}

func root(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	result, err := NewDbHandle(ctx).Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := guestbookTemplate.Execute(w, result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := HostRecord{
		Content: r.FormValue("content"),
		Date:    time.Now(),
	}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	err := NewDbHandle(c).Write(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

var guestbookTemplate = template.Must(template.New("book").Parse(reportTemplateHTML))

const reportTemplateHTML = `
<html>
  <head>
    <title>Go Guestbook</title>
  </head>
  <body>
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
    </form>
  </body>
</html>
`

var signTemplate = template.Must(template.New("sign").Parse(signTemplateHTML))

const signTemplateHTML = `
<html>
  <body>
    <p>-----------------</p>
    <pre>{{.}}</pre>
    <p>-----------------</p>
  </body>
</html>
`

func getUser(w http.ResponseWriter, req *http.Request) *user.User {
	c := appengine.NewContext(req)
	u := user.Current(c)
	if u != nil {
		return u
	}
	url, err := user.LoginURL(c, req.URL.String())
	if err == nil {
		// Redirect to login.
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func form(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, guestbookForm)
}

const guestbookForm = `
<html>
  <body>
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign"></div>
    </form>
  </body>
</html>
`

func show(w http.ResponseWriter, r *http.Request) {
	userName := "unknown"
	if demandLogin {
		u := getUser(w, r)
		if u == nil {
			return
		}
		userName = u.String()
	}
	fmt.Fprintf(w, "Hello %s\n", userName)
	fmt.Fprintf(w, "v23.namespace.root=127.0.0.5:23000\n")
}
