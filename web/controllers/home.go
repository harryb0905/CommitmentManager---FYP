package controllers

import (
	"net/http"
	"html/template"
	"strings"
	"github.com/scc300/scc300-network/blockchain"
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

// Handler to manage the home page
func (app *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
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
  if r.FormValue("submitted") == "true" {
		// Get user input
    fab := app.Fabric
    data.SpecName = r.FormValue("comname")
		comState := r.Form["commitmentState"][0]

		var commitments []blockchain.Commitment
    var err error
    
    spec, er := fab.GetSpec(data.SpecName) // er - could'nt get spec
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
  }
  renderTemplate(w, r, "home.html", data)
}
