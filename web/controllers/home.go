package controllers

import (
	"net/http"
  "encoding/json"
)

type commitment struct {
  ObjectType  string `json:"docType"`      // docType is used to distinguish the various types of objects in state database
  Name        string `json:"name"`         // the fieldtags are needed to keep case from bouncing around
  Owner       string `json:"owner"`        // Owner/creator of the commitment
  DateCreated string `json:"datecreated"`  // Date the commitment was created
  Summary     string `json:"summary"`      // Human-readable string of commitment
  Source      string `json:"source"`       // String to store commitment source code (quark)
}

func (app *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	helloValue, err1 := app.Fabric.QueryHello()
	if err1 != nil {
		http.Error(w, "Unable to query hello on the blockchain", 500)
	}

  data := &struct {
    TransactionId string
    Failed        bool
    Response      bool
    Hello         string
    Name          string
    Owner         string
    DateCreated   string
    Summary       string
    Source        string
  }{
    TransactionId: "",
    Failed:        true,
    Response:      false,
    Hello:         helloValue,
  }
  if r.FormValue("submitted") == "true" {
    comName := r.FormValue("comname")
    com, err := app.Fabric.QueryCommitment(comName)
    if err != nil {
      // http.Error(w, "Unable to query commitment on the blockchain", 500)
      data.Response = true
    } else { 
      s := string(com)
      res := commitment{}
      json.Unmarshal([]byte(s), &res)

      data.TransactionId = com
      data.Failed = false
      data.Response = true
      data.Name = res.Name
      data.Owner = res.Owner
      data.DateCreated = res.DateCreated
      data.Summary = res.Summary
      data.Source = res.Source
    }
  }
  renderTemplate(w, r, "home.html", data)
}


