package controllers

import (
	"net/http"
	"html/template"
	"strings"
	"github.com/chainHero/heroes-service/blockchain"
	c "github.com/chainHero/heroes-service/chaincode/commitments"
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
    TransactionId  string
    Failed         bool
		Response			 bool
    Coms			     []c.QueryResponse
		ComState			 string
		SpecName       string
		SpecSource		 template.HTML
		SpecSummary		 string
  }{
    TransactionId: "",
    Failed:        true,
		Response:			 false,
    Coms:          nil,
		ComState:			 "",
		SpecName:      "",
		SpecSource:		 ``,
		SpecSummary:	 "",
  }
  if r.FormValue("submitted") == "true" {
		// Obtain user input
    comName := r.FormValue("comname")
		comState := r.Form["commitmentState"][0]

		if (comName != "") {
			switch comState {
				case "created":
		      coms, com, err := c.GetCreatedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
						data.Response = true
			    } else {
					  data.TransactionId = comName
			      data.Failed = false
			      data.Response = true
						data.Coms = coms
						data.SpecSource = template.HTML(replacer.Replace(com.Source))
						data.SpecSummary = com.Summary
			    }
					break
		    case "detached":
					coms, com, err := c.GetDetachedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
						data.Response = true
			    } else {
					  data.TransactionId = comName
			      data.Failed = false
			      data.Response = true
						data.Coms = coms
						data.SpecSource = template.HTML(replacer.Replace(com.Source))
						data.SpecSummary = com.Summary
			    }
					break
		    case "expired":
					coms, com, err := c.GetDetachedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
						data.Response = true
			    } else {
					  data.TransactionId = comName
			      data.Failed = false
			      data.Response = true
						data.Coms = coms
						data.SpecSource = template.HTML(replacer.Replace(com.Source))
						data.SpecSummary = com.Summary
			    }
					break
		    case "discharged":
					coms, com, err := c.GetDischargedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
						data.Response = true
			    } else {
					  data.TransactionId = comName
			      data.Failed = false
			      data.Response = true
						data.Coms = coms
						data.SpecSource = template.HTML(replacer.Replace(com.Source))
						data.SpecSummary = com.Summary
			    }
					break
		    case "violated":
					coms, com, err := c.GetViolatedCommitments(comName, app.Fabric)
					if err != nil {
						// http.Error(w, "Unable to query created commitments on the blockchain", 500)
						data.Response = true
			    } else {
					  data.TransactionId = comName
			      data.Failed = false
			      data.Response = true
						data.Coms = coms
						data.SpecSource = template.HTML(replacer.Replace(com.Source))
						data.SpecSummary = com.Summary
			    }
					break
		  }
			data.ComState = strings.Title(comState)
			data.SpecName = comName
		}
  }
  renderTemplate(w, r, "home.html", data)
}
