// Package viewstate defines view state for the index page.
//
// The purpose of the view state is to enable sharing domain models between the
// frontend and the backend. By sharing a state definition, the server can perform
// an initial render and encode its state in a form that the browser can easily
// retrieve.
package viewstate

type ViewState struct {
	Items []string
}
