package controllers

import (
  "fmt"
  "net/http"
  "html/template"
  "strings"
  "time"
  "encoding/json"
  "bytes"
  
  "github.com/satori/go.uuid"
  "github.com/scc300/scc300-network/blockchain"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

const (
  TimeFormat = "Mon Jan _2 15:04:05 2006"
)

// Stores user interface data
type Data struct {
  SpecName        string
  ComState        string
  SpecSource      template.HTML
  Response        bool
  Coms            []blockchain.Commitment
  ParsedSpec      *q.Spec
  NumComs         int
  Failed          bool
  FailMsg         string
  CompilationMsg  string
  CompilationFail bool
}

// Reference to blockchain package
type Application struct {
  Fabric *blockchain.FabricSetup
}

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

// Main handler for both merchants and customers to perform operations
func (app *Application) MainHandler(data *Data, w http.ResponseWriter, r *http.Request) {
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

    // Compile spec to check syntax
    _, er := q.Parse(specContents)
    if er != nil {
      data.CompilationMsg = er.Error()
      data.CompilationFail = true
    } else {
      // Upload new spec to blockchain
      _, err = fab.InvokeInitSpec(specContents)
      if err != nil {
        data.FailMsg = err.Error()
        data.Failed = true
      }
    }
    data.SpecName = r.FormValue("comname")
  }

  // Query all commitments by name and state
  if r.FormValue("query-commitments") == "true" {
    // Get user input
    data.SpecName = r.FormValue("comname")
    comState := r.Form["commitmentState"][0]

    var commitments []blockchain.Commitment
    
    spec, er := fab.GetSpec(data.SpecName)

    if er != nil {
      data.FailMsg = er.Error()
      data.Failed = true
    } else {
      // Obtain commitments based on state (e.g. created, detached, expired, discharged, violated)
      commitments, er = fab.GetCommitments(data.SpecName, comState)
      if er != nil {
        data.FailMsg = er.Error()
        data.Failed = true
      } else {
        // Get event names and argument lists for adding data
        parsedSpec, er := q.Parse(spec.Source)
        if er != nil {
          data.FailMsg = er.Error()
          data.Failed = true
        }
        data.ParsedSpec = parsedSpec
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
  } else if r.FormValue("submitted-data") == "true" {
    // Upload data for a commitment spec event (detach and discharge only)
    r.ParseForm()

    dataMap := map[string]string{
      "docType": r.FormValue("event"),
      "comID": r.FormValue("comID"),
      "date": time.Now().Format(TimeFormat),
    }
    for key, values := range r.Form {
      dataMap[key] = values[0]
    }
    delete(dataMap, "submitted-data");

    // Marshal to JSON and add to blockchain database
    jsonMap, _ := json.Marshal(dataMap)
    _, err := fab.InvokeInitCommitmentData([]string{string(jsonMap)})
    if (err != nil) {
      data.FailMsg = err.Error()
      data.Failed = true
    }
  } else if r.FormValue("submitted-commitment") == "true" {
    // Add new commitment (i.e. a new Offer event)
    r.ParseForm()

    dataMap := map[string]string{
      "docType": r.FormValue("event"),
      "comID": uuid.NewV4().String(),
      "date": time.Now().Format(TimeFormat),
      "debtor": r.FormValue("debtor"),
      "creditor": r.FormValue("creditor"),
    }
    for key, values := range r.Form {
      dataMap[key] = values[0]
    }
    delete(dataMap, "submitted-commitment");

    // Marshal to JSON and add to blockchain database
    jsonMap, _ := json.Marshal(dataMap)
    _, err := fab.InvokeInitCommitmentData([]string{string(jsonMap)})
    if (err != nil) {
      data.FailMsg = err.Error()
      data.Failed = true
    }
  }
}


