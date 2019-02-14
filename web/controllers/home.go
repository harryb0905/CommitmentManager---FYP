package controllers

import (
	"net/http"
	"html/template"
	"strings"
  "fmt"

	"github.com/scc300/scc300-network/blockchain"
)

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
)

type Application struct {
	Fabric *blockchain.FabricSetup
}

type Commitment struct {
  ComID    string
  States []ComState
}

type ComState struct {
  Name  string
  Data  map[string]interface{}
}

type CommitmentMeta struct {
  Name     string  `json:"name"`
  Source   string  `json:"source"`
  Summary  string  `json:"summary"`
}

func (app *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
    TransactionId  		string
    Failed         		bool
		Response			 		bool
		NumComs						int
		Coms 							[]Commitment
		ComState			 		string
		SpecName       		string
		SpecSource		 		template.HTML
		SpecSummary		 		string
  }{
    TransactionId: 		 "",
    Failed:        		 false,
		Response:			 		 false,
		NumComs:					 0,
		Coms:							 nil,
		ComState:			 		 "",
		SpecName:      		 "",
		SpecSource:		 		 ``,
		SpecSummary:	 		 "",
  }
  if r.FormValue("submitted") == "true" {
		// Get user input
    comName := r.FormValue("comname")
		comState := r.Form["commitmentState"][0]

		var commitments []Commitment
		var com CommitmentMeta
		// var err error

    data.ComState = strings.Title(comState)
    data.SpecName = comName

    fab := app.Fabric

    // Obtain commitments based on state
		switch comState {
			case "created":
	      res, er := fab.GetCreatedCommitments(comName)
        fmt.Println("response:",res, er)
				if er != nil {
          data.Failed = true
				}
				break
	   //  case "detached":
				// commitments, com, err = app.GetDetachedCommitments(comName, false, app.Fabric)
				// if err != nil {
    //       data.Failed = true
    //     }
				// break
	   //  case "expired":
				// commitments, com, err = app.GetExpiredCommitments(comName, app.Fabric)
				// if err != nil {
    //       data.Failed = true
    //     }
				// break
	   //  case "discharged":
				// commitments, com, err = app.GetDischargedCommitments(comName, false, app.Fabric)
				// if err != nil {
    //       data.Failed = true
    //     }
				// break
	   //  case "violated":
				// commitments, com, err = app.GetViolatedCommitments(comName, app.Fabric)
				// if err != nil {
    //       data.Failed = true
    //     }
				// break
	  }
    if (!data.Failed) {
			data.TransactionId = comName
			data.Response = true
			data.NumComs = len(commitments)
			data.Coms = commitments
			data.SpecSource = template.HTML(replacer.Replace(com.Source))
			data.SpecSummary = com.Summary
    }
  }
  renderTemplate(w, r, "home.html", data)
}
