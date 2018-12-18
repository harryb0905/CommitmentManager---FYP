package main

import (
	"fmt"
	"github.com/chainHero/heroes-service/blockchain"
	"github.com/chainHero/heroes-service/web"
	"github.com/chainHero/heroes-service/web/controllers"
  "encoding/json"
	"os"
)

type Record struct {
  DateCreated string `json:"datecreated"`
  DocType string `json:"docType"`
  Name string `json:"name"`
  Owner string `json:"owner"`
  Source string `json:"source"`
  Summary string `json:"summary"`
}

type Result struct {
  Key     string `json:"key"`
  Record *Record `json:"record"`
}

func main() {
	// Definition of the Fabric SDK properties
	fSetup := blockchain.FabricSetup{
		// Network parameters
		OrdererID: "orderer.hf.chainhero.io",

		// Channel parameters
		ChannelID:     "chainhero",
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/chainHero/heroes-service/fixtures/artifacts/chainhero.channel.tx",

		// Chaincode parameters
		ChainCodeID:     "heroes-service",
		ChaincodeGoPath: os.Getenv("GOPATH"),
		ChaincodePath:   "github.com/chainHero/heroes-service/chaincode/",
		OrgAdmin:        "Admin",
		OrgName:         "org1",
		ConfigFile:      "config.yaml",

		// User parameters
		UserName: "User1",
	}

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
		return
	}
	// Close SDK
	defer fSetup.CloseSDK()

	// Install and instantiate the chaincode
	err = fSetup.InstallAndInstantiateCC()
	if err != nil {
		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
		return
	}

  // Init a commitment on chaincode
  args := []string{"SellItem", "Harry Baines", "22/11/18", "If SellItem is blah blah blah...", "spec SellItem dID to cID"}
  response, err := fSetup.InvokeInitCommitment(args)
  if err != nil {
    fmt.Printf("Unable to initialise commitment on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment initialisation: %s\n", response)
  }

	// Put dummy data on blockchain corresponding to a spec (because we assume data already exists)
	// Dummy data 1
	args = []string{"Offer", "Chair", "10.99", "Good"}
	response, err = fSetup.InvokeInitCommitmentData(args)
  if err != nil {
    fmt.Printf("Unable to initialise commitment data on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment data initialisation: %s\n", response)
  }

	// Dummy data 2
	args = []string{"Offer", "Table", "29.99", "Slightly Damaged"}
	response, err = fSetup.InvokeInitCommitmentData(args)
  if err != nil {
    fmt.Printf("Unable to initialise commitment data on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment data initialisation: %s\n", response)
  }

  // Init another commitment on chaincode
  args = []string{"Refund", "Harry Baines", "10/05/18", "If Refund is blah blah blah...", "spec Refund dID to cID"}
  response, err = fSetup.InvokeInitCommitment(args)
  if err != nil {
    fmt.Printf("Unable to initialise commitment on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment initialisation: %s\n", response)
  }

  // Query a commitment on chaincode - match against existing data
  response, err = fSetup.QueryCommitment("SellItem")
  if err != nil {
    fmt.Printf("Unable to query commitment on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from the commitment query: %s\n", response)
  }

  // Perform paramterised query on chaincode
  query := fmt.Sprintf("{\"selector\":{\"docType\":\"commitment\",\"name\":\"%s\"}}", "SellItem")
  response, err = fSetup.RichQuery(query)
  if err != nil {
    fmt.Printf("Unable to perform rich query on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from the rich query: %s\n", response)

    // Unmarshal JSON
    results := []Result{}
    err := json.Unmarshal([]byte(response), &results)
    if err != nil {
      fmt.Println(err)
    } else {
      for _, res := range results {
        fmt.Println(res.Record.Name, res.Record.Owner, res.Record.Summary)
      }
    }
  }

  // Query 1: Get SellItem created queries (make a QueryCreated method for this)
	query = fmt.Sprintf("{\"selector\":{\"docType\":\"Table\",\"name\":\"%s\"}}", "Offer")
  response, err = fSetup.RichQuery(query)
  if err != nil {
    fmt.Printf("Unable to perform rich query on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from the rich query: %s\n", response)

    // Unmarshal JSON
    results := []Result{}
    err = json.Unmarshal([]byte(response), &results)
    if err != nil {
      fmt.Println(err)
    } else {
      fmt.Println(results)
      for _, res := range results {
        fmt.Println(res.Record)
      }
    }
  }

	// Launch the web application listening
	app := &controllers.Application{
		Fabric: &fSetup,
	}
	web.Serve(app)
}
