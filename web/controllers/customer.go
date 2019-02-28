package controllers

import (
  "fmt"
	"net/http"
	"html/template"
	"strings"
  "time"
  "encoding/json"
  "bytes"
	
  "github.com/scc300/scc300-network/blockchain"
  "github.com/satori/go.uuid"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

// Syntax highlighting for quark language
var replacer = strings.NewReplacer(
	"\n", "<br>",
	"\t", "&emsp;",
	"spec",
	`<span style="color:red;">spec</span>`,
	"to",
	`<span style="color:red;">to</span>`,
	"create",
	`<span style="color:blue;">create</span>`,
	"detach",
	`<span style="color:blue;">detach</span>`,
	"discharge",
	`<span style="color:blue;">discharge</span>`,
  "deadline",
  `<span style="font-style:italic;">deadline</span>`,
)

type Application struct {
	Fabric *blockchain.FabricSetup
}

// Handler to manage the customer application
func (app *Application) CustomerHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		SpecName       		string
    ComState          string
		SpecSource		 		template.HTML
    Response          bool
    Coms              []blockchain.Commitment
    NumComs           int
    Failed            bool
  }{
		SpecName:      		 "",
    ComState:          "",
		SpecSource:		 		 ``,
    Response:          false,
    Coms:              nil,
    NumComs:           0,
    Failed:            false,
  }
  fab := app.Fabric

  // Get spec file upload
  if r.Method == "POST" {
    file, _, err := r.FormFile("uploadfile")
    if err != nil {
        fmt.Println(err)
        return
    }
    var buff bytes.Buffer
    buff.ReadFrom(file)
    specContents := buff.String()
    defer file.Close()

    // Upload new spec to blockchain
    _, err = fab.InvokeInitSpec(specContents)
    if err != nil {
      data.Failed = true
    }
    data.SpecName = r.FormValue("comname")
  }


  if r.FormValue("query-commitments") == "true" {
		// Get user input
    data.SpecName = r.FormValue("comname")
    comState := r.Form["commitmentState"][0]

    var commitments []blockchain.Commitment
    var err error
    
    spec, er := fab.GetSpec(data.SpecName)
    if er != nil {
      data.Failed = true
    } else {
      // Obtain commitments based on state (e.g. created, detached, expired, discharged, violated)
      commitments, err = fab.GetCommitments(data.SpecName, comState)
      if err != nil {
        data.Failed = true
      }
    }
    // Prepare user interface output data
    if (!data.Failed) {
      data.ComState = strings.Title(comState)
      data.SpecSource = template.HTML(replacer.Replace(spec.Source))
      data.Response = true
      data.Coms = commitments
      data.NumComs = len(commitments)
    }
  } else if r.FormValue("query-spec") == "true" {
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
  renderTemplate(w, r, "customer.html", data)
}
