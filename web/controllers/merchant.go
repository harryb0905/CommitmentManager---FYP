package controllers

import (
	"net/http"
	// q "github.com/chainHero/heroes-service/quark"
)

// Handler to manage the merchant page
func (app *Application) MerchantHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		TransactionId string
		Success       bool
		Response      bool
	}{
		TransactionId: "",
		Success:       false,
		Response:      false,
	}
	renderTemplate(w, r, "home.html", data)
}
