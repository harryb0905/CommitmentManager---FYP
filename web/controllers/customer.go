package controllers

import (
  "net/http"
)

// Handler to manage the customer application
func (app *Application) CustomerHandler(w http.ResponseWriter, r *http.Request) {
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
  renderTemplate(w, r, "customer.html", data)
}
