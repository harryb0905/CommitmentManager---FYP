package controllers

import (
  "fmt"
	"net/http"
  "html/template"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

// Handler to manage the merchant application
func (app *Application) MerchantHandler(w http.ResponseWriter, r *http.Request) {
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
  fab := app.Fabric
  if r.FormValue("submitted") == "true" {
    // Get user input
    data.SpecName = r.FormValue("specname")

    // Obtain spec source, parse and prepare user interface output data
    spec, err := fab.GetSpec(data.SpecName)  

    if err != nil {
      data.Failed = true
    } else {
      _, er := q.Parse(spec.Source)
      if er != nil {
        data.Failed = true
      }
      data.SpecSource = template.HTML(replacer.Replace(spec.Source))
      data.Response = true
    }
  } else if r.FormValue("submitted-data") == "true" {
    item := r.FormValue("item")
    // fmt.Println("Item: " + item)

    // jsonStrs := make([]string, 0)

    // jsonStrs = append(jsonStrs, item)

    // fab.InvokeInitCommitmentData(jsonStrs)
  }
	renderTemplate(w, r, "merchant.html", data)
}
