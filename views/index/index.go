// +build js

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"html/template"

	"github.com/albrow/vdom"
	"honnef.co/go/js/dom"

	"github.com/dradtke/isomorphic-golang/views/index/viewstate"
)

var (
	state  viewstate.ViewState
	t      *template.Template
	listEl *dom.HTMLDivElement
	tree   *vdom.Tree
)

func main() {
	var (
		err        error
		templateEl = dom.GetWindow().Document().QuerySelector(`script[type="text/template"][data-tmpl="index/index.tmpl"]`).(*dom.HTMLScriptElement)
		stateEl    = dom.GetWindow().Document().QuerySelector(`script[type="application/gob"][data-for="list"]`).(*dom.HTMLScriptElement)
	)
	listEl = dom.GetWindow().Document().GetElementByID("list").(*dom.HTMLDivElement)

	if t, err = template.New("").Delims("[[", "]]").Parse(templateEl.Text); err != nil {
		panic(err)
	}

	rawData, err := base64.StdEncoding.DecodeString(stateEl.Text)
	if err != nil {
		panic(err)
	}
	if err := gob.NewDecoder(bytes.NewReader(rawData)).Decode(&state); err != nil {
		panic(err)
	}

	tree, err = vdom.Parse([]byte(listEl.InnerHTML()))
	if err != nil {
		panic(err)
	}

	// Handle new items.
	dom.GetWindow().Document().QuerySelector("#newItemForm").AddEventListener("submit", false, newItemListener)
	if err := render(); err != nil {
		panic(err)
	}
}

func newItemListener(e dom.Event) {
	e.PreventDefault()

	newValue := e.Target().QuerySelector("input").(*dom.HTMLInputElement).Value
	state.Items = append(state.Items, newValue)

	if err := render(); err != nil {
		panic(err)
	}
}

// render uses vdom to perform a React-like virtual DOM render.
func render() error {
	var buf bytes.Buffer
	if err := t.Execute(&buf, state); err != nil {
		return err
	}
	newTree, err := vdom.Parse(buf.Bytes())
	if err != nil {
		return err
	}
	patches, err := vdom.Diff(tree, newTree)
	if err != nil {
		return err
	}
	if err := patches.Patch(listEl); err != nil {
		return err
	}
	tree = newTree
	return nil
}
