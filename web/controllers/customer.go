package controllers

import (
	"net/http"
  "html/template"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

// Handler to manage the merchant page
func (app *Application) CustomerHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
    SpecName     string
    SpecSource   template.HTML
		Failed       bool
		Response     bool
	}{
    SpecName:    "",
    SpecSource:  ``,
		Failed:      false,
		Response:    false,
	}
  if r.FormValue("submitted") == "true" {
    // Get user input
    fab := app.Fabric
    data.SpecName = r.FormValue("specname")

    // Obtain spec source, parse and prepare user interface output data
    spec, err := fab.GetSpec(data.SpecName)  
    parsedSpec := q.Parse(spec.Source)

    if err != nil {
      data.Failed = true
    } else {
      data.SpecSource = template.HTML(replacer.Replace(spec.Source))
      data.Response = true
    }
  }
	renderTemplate(w, r, "customer.html", data)
}
