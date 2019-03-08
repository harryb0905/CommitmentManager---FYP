package main

import (
	"fmt"
	"os"
  "io/ioutil"
  "encoding/json"
  "reflect"
  "log"

	"github.com/scc300/scc300-network/blockchain"
  "github.com/scc300/scc300-network/web"
  "github.com/scc300/scc300-network/web/controllers"
)

// Blockchain initialization and start customer and merchant web applications
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

  // Commitment initialisation - Get spec source from file and initialise
  specSource := getSpecSource("./specs/SellItem.quark")
  _, err = fSetup.InvokeInitSpec(specSource)
  if err != nil {
    log.Fatalf("Unable to initialise SellItem commitment on the chaincode: %v\n", err)
  }

  // Commitment Data Initialisation - Read JSON file and add initial data to blockchain (because we assume data already exists)
  jsonStrs := getJSONObjectStrsFromFile("./specs/test_data.json")
	_, err = fSetup.InvokeInitCommitmentData(jsonStrs)
  if err != nil {
    log.Fatalf("Unable to initialise commitment data on the chaincode: %v\n", err)
  }

  // Launch the customer web application
  // go func() {
  //   customer.Serve(&controllers.Application{
  //     Fabric: &fSetup,
  //   })
  // }()

  // // Launch the merchant web application
  // merchant.Serve(&controllers.Application{
  //   Fabric: &fSetup,
  // })

  // Create 2 servers - 1 merchant, 1 customer
  web.StartServers(&controllers.Application{
    Fabric: &fSetup,
  })
}

// Function to obtain the specification source code as a string (input is a filepath to the .quark file)
func getSpecSource(filepath string) (source string) {
  data, err := ioutil.ReadFile(filepath)
  if (err != nil) {
    log.Fatalf("Couldn't read spec file %s", filepath)
  }
  return string(data)
}

// Obtains JSON strings of data from a given filepath as a string
// Returns a slice of strings - each a JSON object in string form
func getJSONObjectStrsFromFile(filepath string) (strs []string) {
  data, err := ioutil.ReadFile(filepath)
  if (err != nil) {
    log.Fatalf("Couldn't read JSON file %s", filepath)
  }

  // Parse the JSON
  var objs interface{}
  json.Unmarshal([]byte(string(data)), &objs)

  // Ensure that it is an array of objects.
  objArr, ok := objs.([]interface{})
  if !ok {
    log.Fatal("expected an array of objects")
  }

  // Handle each object as a map[string]interface{}
  jsonStrs := make([]string, 0)
  for i, obj := range objArr {
    obj, ok := obj.(map[string]interface{})
    if !ok {
      log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(objArr[i]))
    }
    jsonString, _ := json.Marshal(obj)
    jsonStrs = append(jsonStrs, string(jsonString))
  }
  return jsonStrs
}
