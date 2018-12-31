package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	index_viewstate "github.com/dradtke/isomorphic-golang/views/index/viewstate"
)

var viewFuncs = template.FuncMap{
	"render": func(t *template.Template, serverState interface{}, divId string, clientState interface{}, templates ...string) (template.HTML, error) {
		var rendered, buf bytes.Buffer

		var cts *template.Template
		for _, name := range templates {
			buf.Reset()
			var err error
			if err = t.ExecuteTemplate(&buf, name, serverState); err != nil {
				return template.HTML(""), fmt.Errorf("failed to execute client template %s: %s", name, err)
			}
			if cts == nil {
				cts, err = template.New("").Delims("[[", "]]").Parse(buf.String())
			} else {
				cts, err = cts.New(name).Delims("[[", "]]").Parse(buf.String())
			}
			if err != nil {
				return template.HTML(""), fmt.Errorf("failed to parse client template %s: %s", name, err)
			}
			// Write the client template text to a <script> tag that can be read and used by the frontend.
			rendered.WriteString(fmt.Sprintf(`<script type="text/template" data-tmpl="%s">%s</script>`, name, buf.String()))
		}

		// Initial server-side render of the defined client-side templates.
		rendered.WriteString(fmt.Sprintf(`<div id="%s">`, divId))
		if err := cts.Execute(&rendered, clientState); err != nil {
			return template.HTML(""), errors.New("failed to execute client templates: " + err.Error())
		}
		rendered.WriteString(`</div>`)

		// Write the initial state as a gob that can be picked up by the frontend.
		buf.Reset()
		if err := gob.NewEncoder(&buf).Encode(clientState); err != nil {
			return template.HTML(""), errors.New("failed to encode client-side state: " + err.Error())
		}
		rendered.WriteString(fmt.Sprintf(`<script type="application/gob" data-for="%s">%s</script>`, divId, base64.StdEncoding.EncodeToString(buf.Bytes())))

		return template.HTML(rendered.String()), nil
	},
}

// ParseViews locates and parses server-side views
func ParseViews() *template.Template {
	return parse(".html")
}

// ParseTemplates locates and parses client-side templates.
func ParseTemplates() *template.Template {
	return parse(".tmpl")
}

func parse(extension string) *template.Template {
	const viewsDir = "views"
	t := template.New("")
	if err := filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(info.Name(), extension) {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		name := path[len(viewsDir)+1:] // trim the views dir from the template name
		_, err = t.New(name).Funcs(viewFuncs).Parse(string(data))
		return err
	}); err != nil {
		log.Fatal(err)
	}
	return t
}

// IndexHandler renders the index page.
func IndexHandler(views, templates *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Templates    *template.Template
			InitialState index_viewstate.ViewState
		}{
			Templates: templates,
			InitialState: index_viewstate.ViewState{
				Items: []string{"First Item", "Second Item", "Third Item"},
			},
		}

		if err := views.ExecuteTemplate(w, "index/index.html", data); err != nil {
			log.Printf("error executing template: %s", err)
		}
	})
}

func main() {
	views := ParseViews()
	templates := ParseTemplates()

	http.Handle("/", IndexHandler(views, templates))
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
