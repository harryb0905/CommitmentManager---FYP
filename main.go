package main

import (
	"fmt"
	"time"
	"os"

	"github.com/satori/go.uuid"
	"github.com/scc300/scc300-network/blockchain"
	"github.com/scc300/scc300-network/web"
	"github.com/scc300/scc300-network/web/controllers"
)

const (
	TimeFormat = "Mon Jan _2 15:04:05 2006"
)

type Commitment struct {
  DateCreated string 	`json:"datecreated"`
  DocType 		string 	`json:"docType"`
  Name 				string 	`json:"name"`
  Owner 			string 	`json:"owner"`
  Source 			string 	`json:"source"`
  Summary 		string 	`json:"summary"`
}

type Result struct {
  Key     		string 		 `json:"key"`
  Commitment *Commitment `json:"record"`
}

type QueryResponse struct {
  Key        string
  Value      string
  Namespace  string
}

func main() {
	// Definition of the Fabric SDK properties
	fSetup := blockchain.FabricSetup{
		// Network parameters
		OrdererID: "orderer.hf.scc300.io",

		// Channel parameters
		ChannelID:     "scc300",
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/scc300/scc300-network/fixtures/artifacts/scc300.channel.tx",

		// Chaincode parameters
		ChainCodeID:     "scc300-network",
		ChaincodeGoPath: os.Getenv("GOPATH"),
		ChaincodePath:   "github.com/scc300/scc300-network/chaincode/",
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
  args := []string{
		"SellItem",
		"Harry Baines",
		"22/11/18",
		`spec SellItem dID to cID
	  	create Offer [item,price,quality]
	  	detach Pay [amount,address,shippingtype,deadline=5]
	  	discharge Delivery [deadline=5]`,
	}

  response, err := fSetup.InvokeInitCommitment(args)
  if err != nil {
    fmt.Printf("Unable to initialise commitment on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment initialisation: %s\n", response)
  }

	// Add initial data to blockchain (because we assume data already exists)
	ID1 := genUUIDv4().String()
	ID2 := genUUIDv4().String()
	ID3 := genUUIDv4().String()

	jsonStrs := []string{
	  `{
			"docType": "Offer",
			"comID": "` + ID1 + `",
			"debtor": "Harry",
			"creditor": "John",
			"item": "Chair",
			"price": "10.99",
			"quality": "Good",
			"date": "` + time.Date(2018, time.December, 20, 18, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
	   }`,
	  `{
			"docType": "Offer",
			"comID": "` + ID2 + `",
			"debtor": "Yash",
			"creditor": "Georgi",
			"item": "Lamp",
			"price": "29.99",
			"quality": "Slightly Damaged",
			"date": "` + time.Date(2018, time.December, 20, 20, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
		 }`,
	  `{
			"docType": "Offer",
			"comID": "` + ID3 + `",
			"debtor": "Simon",
			"creditor": "Joe",
			"item": "Beer",
			"price": "9.99",
			"quality": "Good",
			"date": "` + time.Date(2018, time.December, 20, 23, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
		 }`,
		`{
 			"docType": "Pay",
			"comID": "` + ID1 + `",
			"debtor": "Harry",
			"creditor": "John",
 			"amount": "10.99",
 			"address": "49 Garstang Road West",
 			"shippingtype": "Express Delivery",
 			"date": "` + time.Date(2018, time.December, 22, 20, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
 		 }`,
		`{
			"docType": "Pay",
			"comID": "` + ID2 + `",
			"debtor": "Yash",
			"creditor": "Georgi",
			"amount": "29.99",
			"address": "87 Hardhorn Road",
			"shippingtype": "Nominated Day Delivery",
			"date": "` + time.Date(2018, time.December, 24, 20, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
		 }`,
		`{
 			"docType": "Delivery",
 			"comID": "` + ID1 + `",
 			"debtor": "Harry",
 			"creditor": "John",
 			"date": "` + time.Date(2018, time.December, 24, 20, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
 		 }`,
	  `{
		  "docType": "Delivery",
			"comID": "` + ID2 + `",
			"debtor": "Yash",
			"creditor": "Georgi",
			"date": "` + time.Date(2018, time.January, 10, 10, 0, 0, 0, time.UTC).Format(TimeFormat) + `"
		 }`,
	}
	response, err = fSetup.InvokeInitCommitmentData(jsonStrs)
  if err != nil {
    fmt.Printf("Unable to initialise commitment data on the chaincode: %v\n", err)
  } else {
    fmt.Printf("Response from commitment data initialisation: %s\n", response)
  }

  // Init another commitment on chaincode
	// args = []string{
	// 	"Refund",
	// 	"Harry Baines",
	// 	"10/05/18",
	// 	`spec Refund dID to cID
	//   	create Offer [item,price,quality]
	//   	detach Pay [amount,address,shippingtype,deadline=10]
	//   	discharge Refund [deadline=2]`,
	// }
	//
  // response, err = fSetup.InvokeInitCommitment(args)
  // if err != nil {
  //   fmt.Printf("Unable to initialise commitment on the chaincode: %v\n", err)
  // } else {
  //   fmt.Printf("Response from commitment initialisation: %s\n", response)
  // }

  // Query a commitment on chaincode - match against existing data
  // response, err = fSetup.QueryCommitment("SellItem")
  // if err != nil {
  //   fmt.Printf("Unable to query commitment on the chaincode: %v\n", err)
  // } else {
  //   fmt.Printf("Response from the commitment query: %s\n", response)
  // }

  // Perform parameterised query on chaincode
  // query := fmt.Sprintf("{\"selector\":{\"docType\":\"commitment\",\"name\":\"%s\"}}", "SellItem")
  // response, err = fSetup.RichQuery(query)
  // if err != nil {
  //   fmt.Printf("Unable to perform rich query on the chaincode: %v\n", err)
  // } else {
  //   fmt.Printf("Response from the rich query: %s\n", response)
	//
  //   // Unmarshal JSON
  //   results := []Result{}
  //   err := json.Unmarshal([]byte(response), &results)
  //   if err != nil {
  //     p(err)
  //   } else {
  //     for _, res := range results {
  //       p(res.Commitment.Name, res.Commitment.Owner, res.Commitment.Summary)
  //     }
  //   }
  // }

  // Query 1: Get SellItem created queries (make a QueryCreated method for this)
	// query = fmt.Sprintf("{\"selector\":{\"event\":\"%s\"}}", "Offer")
  // response, err = fSetup.RichQuery(query)
  // if err != nil {
  //   fmt.Printf("Unable to perform rich query on the chaincode: %v\n", err)
  // } else {
  //   fmt.Printf("Response from the rich query: %s\n", response)
	//
  //   // Unmarshal JSON
  //   results := []QueryResponse{}
  //   err = json.Unmarshal([]byte(response), &results)
  //   if err != nil {
  //     p(err)
  //   } else {
  //     p(results)
  //     for _, res := range results {
  //       p(res.Value)
  //     }
  //   }
  // }

	// Launch the web application
	app := &controllers.Application{
		Fabric: &fSetup,
	}
	web.Serve(app)
}

func genUUIDv4() (uuid.UUID) {
  return uuid.NewV4()
}
