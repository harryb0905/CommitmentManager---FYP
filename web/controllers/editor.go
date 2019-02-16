package controllers

import (
	"net/http"
	// q "github.com/chainHero/heroes-service/quark"
)

func (app *Application) EditorHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		TransactionId string
		Success       bool
		Response      bool
	}{
		TransactionId: "",
		Success:       false,
		Response:      false,
	}
	renderTemplate(w, r, "editor.html", data)
}
