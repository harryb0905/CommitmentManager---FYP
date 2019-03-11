package controllers

import (
  "net/http"
)

// Handler to manage the merchant application
func (app *Application) MerchantHandler(w http.ResponseWriter, r *http.Request) {
  data := Data{
    SpecName:     "",
    ComState:     "",
    SpecSource:   ``,
    Response:     false,
    Coms:         nil,
    NumComs:      0,
    Failed:       false,
  }
  app.MainHandler(&data, w , r)
  renderTemplate(w, r, "merchant.html", data)
}