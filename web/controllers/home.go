package controllers

import (
	"net/http"
	"html/template"
	"strings"

	"github.com/scc300/scc300-network/blockchain"
	c "github.com/scc300/scc300-network/chaincode/commitments"
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

func (app *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := &struct {
    TransactionId  		string
    Failed         		bool
		Response			 		bool
		NumComs						int
		Coms 							[]c.Commitment
		ComState			 		string
		SpecName       		string
		SpecSource		 		template.HTML
		SpecSummary		 		string
  }{
    TransactionId: 		 "",
    Failed:        		 true,
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

		if (comName != "") {
			var commitments []c.Commitment
			var com c.CommitmentMeta
			var err error

			switch comState {
				case "created":
		      commitments, com, err = c.GetCreatedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
					}
					break
		    case "detached":
					commitments, com, err = c.GetDetachedCommitments(comName, false, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query detached commitments on the blockchain", 500)
					}
					break
		    case "expired":
					commitments, com, err = c.GetExpiredCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query expired commitments on the blockchain", 500)
					}
					break
		    case "discharged":
					commitments, com, err = c.GetDischargedCommitments(comName, false, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query expired commitments on the blockchain", 500)
					}
					break
		    case "violated":
					commitments, com, err = c.GetViolatedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query expired commitments on the blockchain", 500)
					}
					break
		  }

			data.TransactionId = comName
			data.Failed = false
			data.Response = true
			data.NumComs = len(commitments)
			data.Coms = commitments
			data.SpecSource = template.HTML(replacer.Replace(com.Source))
			data.SpecSummary = com.Summary

			data.ComState = strings.Title(comState)
			data.SpecName = comName
		}
  }
  renderTemplate(w, r, "home.html", data)
}
