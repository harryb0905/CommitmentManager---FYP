package controllers

import (
	"net/http"
	"html/template"
	"strings"
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
		Coms 							[]blockchain.Commitment
		ComState			 		string
		SpecName       		string
		SpecSource		 		template.HTML
  }{
    TransactionId: 		 "",
    Failed:        		 false,
		Response:			 		 false,
		NumComs:					 0,
		Coms:							 nil,
		ComState:			 		 "",
		SpecName:      		 "",
		SpecSource:		 		 ``,
  }
  if r.FormValue("submitted") == "true" {
		// Get user input
    fab := app.Fabric
    comName := r.FormValue("comname")
		comState := r.Form["commitmentState"][0]

		var commitments []blockchain.Commitment
    var err error
    
    spec, _ := fab.GetSpec(comName) // er - could'nt get

    data.ComState = strings.Title(comState)
    data.SpecName = comName

    // Obtain commitments based on state
		switch comState {
			case "created":
	      commitments, err = fab.GetCreatedCommitments(comName)
				if err != nil {
          data.Failed = true
				}
        data.Coms = commitments
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
			data.SpecSource = template.HTML(replacer.Replace(spec.Source))
    }
  }
  renderTemplate(w, r, "home.html", data)
}
