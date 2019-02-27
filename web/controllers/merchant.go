package controllers

import (
  "time"
  "encoding/json"
	"net/http"
  "html/template"
  "github.com/satori/go.uuid"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

const (
  TimeFormat = "Mon Jan _2 15:04:05 2006"
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
    debtor := r.FormValue("debtor")
    creditor := r.FormValue("creditor")
    item := r.FormValue("item")
    price := r.FormValue("price")
    quality := r.FormValue("quality")

    dataMap := map[string]string{
      "docType": "Offer", 
      "comID": uuid.NewV4().String(),
      "debtor": debtor,
      "creditor": creditor,
      "item": item, 
      "price": price, 
      "quality": quality,
      "date": time.Now().Format(TimeFormat),
    }
    jsonMap, _ := json.Marshal(dataMap)
    fab.InvokeInitCommitmentData([]string{string(jsonMap)})
  }
	renderTemplate(w, r, "merchant.html", data)
}
