package main

import (
	"fmt"
	"os"
  "io/ioutil"
  "encoding/json"
  "reflect"
  "log"

  q "github.com/scc300/scc300-network/quark"
	"github.com/scc300/scc300-network/blockchain"
	"github.com/scc300/scc300-network/web"
	"github.com/scc300/scc300-network/web/controllers"
)

const (
  JSONDataFile = "./specs/test_data.json"
  GreenTick = "\033[92m" + "\u2713" + "\033[0m"
)

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
  specArgs := compileSpec("./specs/SellItem.quark")
  _, err = fSetup.InvokeInitSpec(specArgs)
  if err != nil {
    log.Fatalf("Unable to initialise commitment on the chaincode: %v\n", err)
  }

  // Commitment Data Initialisation - Read JSON file and add initial data to blockchain (because we assume data already exists)
  jsonStrs := getJSONObjectStrsFromFile(JSONDataFile)
	_, err = fSetup.InvokeInitCommitmentData(jsonStrs)
  if err != nil {
    log.Fatalf("Unable to initialise commitment data on the chaincode: %v\n", err)
  }

	// Launch the web application
	app := &controllers.Application{
		Fabric: &fSetup,
	}
	web.Serve(app)
}

// Function to compile spec file and create list of args for initialisation on blockchain
func compileSpec(filepath string) ([]string) {
  data, err := ioutil.ReadFile(filepath)
  if (err != nil) {
    log.Fatalf("Couldn't read spec file %s", filepath)
  }
  source := string(data)

  spec, err := q.Parse(source)
  if (err != nil) {
    log.Fatal("\nSyntax Error:\n", err, "\n")
  } else {
    fmt.Printf("\n%s spec compiled successfully %s \n\n", spec.Constraint.Name, GreenTick)
  }

  // Create list of args to initialise a new spec
  specArgs := []string{spec.Constraint.Name, source}
  return specArgs
}

func getJSONObjectStrsFromFile(filepath string) (strs []string) {
  data, err := ioutil.ReadFile(filepath)
  if (err != nil) {
    log.Fatalf("Couldn't read JSON file %s", filepath)
  }

  // Parse the JSON - or use json.Decoder.Decode(...)
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
