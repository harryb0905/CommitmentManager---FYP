package main

import (
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  "github.com/satori/go.uuid"
  pb "github.com/hyperledger/fabric/protos/peer"

  "encoding/json"
  "bytes"
)

// HeroesServiceChaincode implementation of Chaincode
type HeroesServiceChaincode struct {
}

type commitment struct {
  ObjectType  string `json:"docType"`     // docType is used to distinguish the various types of objects in state database
  Name        string `json:"name"`        // the field tags are needed to keep case from bouncing around
  Owner       string `json:"owner"`       // Owner/creator of the commitment
  DateCreated string `json:"datecreated"` // Date the commitment was created
  Summary     string `json:"summary"`     // Human-readable string of commitment
  Source      string `json:"source"`      // String to store commitment source code (quark)
}

// ============================================================
// Init - This function is called only one when the chaincode is instantiated.
// So the goal is to prepare the ledger to handle future requests.
// ============================================================
func (t *HeroesServiceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
  fmt.Println("########### HeroesServiceChaincode Init ###########")

  // Get the function and arguments from the request
  function, _ := stub.GetFunctionAndParameters()

  // Check if the request is the init function
  if function != "init" {
    return shim.Error("Unknown function call")
  }

  // Return a successful message
  return shim.Success(nil)
}

// ============================================================
// Invoke - All future invoke requests will arrive here
// ============================================================
func (t *HeroesServiceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
  fmt.Println("########### HeroesServiceChaincode Invoke ###########")

  // Get the function and arguments from the request
  function, args := stub.GetFunctionAndParameters()

  // Check whether the number of arguments is sufficient
  if len(args) < 1 {
    return shim.Error("The number of arguments is insufficient.")
  }

  // Handle different functions
  if function == "initCommitment" {
    return t.initCommitment(stub, args)
  } else if function == "initCommitmentData" {
    return t.initCommitmentData(stub, args)
  } else if function == "readCommitment" {
    return t.readCommitment(stub, args)
  } else if function == "invoke" {
    return t.invoke(stub, args)
  } else if function == "query" {
    return t.query(stub, args)
  } else if function == "richQuery" {
    return t.richQuery(stub, args)
  }

  // If the arguments given don’t match any function, we return an error
  return shim.Error("Unknown action, check the first argument")
}

// ====================================================================
// initCommitment - create a new commitment, store into chaincode state
// ====================================================================
func (t *HeroesServiceChaincode) initCommitment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var err error

  //   0                1                      2                                     3
  // "SellItem", "HarryBaines", "If a SellItem is blah blah blah...", "commitment SellItem dID to cID..."
  if len(args) != 5 {
    return shim.Error("Incorrect number of arguments. Expecting 5")
  }

  // ==== Input sanitation ====
  fmt.Println("- start init commitment")
  if len(args[0]) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }
  if len(args[1]) <= 0 {
    return shim.Error("2nd argument must be a non-empty string")
  }
  if len(args[2]) <= 0 {
    return shim.Error("3rd argument must be a non-empty string")
  }
  if len(args[3]) <= 0 {
    return shim.Error("4th argument must be a non-empty string")
  }
  if len(args[4]) <= 0 {
    return shim.Error("5th argument must be a non-empty string")
  }
  commitmentName := args[0]
  owner := args[1]
  datecreated := args[2]
  summary := args[3]
  source := args[4]

  // ==== Check if commitment already exists ====
  commitmentAsBytes, err := stub.GetState(commitmentName)
  if err != nil {
    return shim.Error("Failed to get commitment: " + err.Error())
  } else if commitmentAsBytes != nil {
    fmt.Println("This commitment already exists: " + commitmentName)
    return shim.Error("This commitment already exists: " + commitmentName)
  }

  // ==== Create commitment object and marshal to JSON ====
  objectType := "commitment"
  commitment := &commitment{objectType, commitmentName, owner, datecreated, summary, source}
  commitmentJSONasBytes, err := json.Marshal(commitment)
  if err != nil {
    return shim.Error(err.Error())
  }

  // === Save commitment to state ===
  err = stub.PutState(commitmentName, commitmentJSONasBytes)
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Index the commitment to enable color-based range queries, e.g. return all blue commitments ====
  //  An 'index' is a normal key/value entry in state.
  //  The key is a composite key, with the elements that you want to range query on listed first.
  //  In our case, the composite key is based on indexName~owner~name.
  //  This will enable very efficient state range queries based on composite keys matching indexName~color~*
  indexName := "owner~name"
  ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{commitment.Owner, commitment.Name})
  if err != nil {
    return shim.Error(err.Error())
  }

  //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the commitment.
  //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
  value := []byte{0x00}
  stub.PutState(ownerNameIndexKey, value)

  // ==== Commitment saved and indexed. Return success ====
  fmt.Println("- end init commitment")

  // Notify listeners that an event "eventInvoke" have been executed (check line 24 in the file invoke.go)
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(nil)
}

// ======================================================================
// initCommitmentData - adds commitment data to blockchain to be queried
// ======================================================================
func (t *HeroesServiceChaincode) initCommitmentData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var err error

  //   0         1       2       3       4      5
  // "Offer", "Harry", "John", "Chair", "10", "Good"
  if len(args) != 6 {
    return shim.Error("Incorrect number of arguments. Expecting 6")
  }

  // ==== Input sanitation ====
  fmt.Println("- start init commitment")
  if len(args[0]) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }
  if len(args[1]) <= 0 {
    return shim.Error("2nd argument must be a non-empty string")
  }
  if len(args[2]) <= 0 {
    return shim.Error("3rd argument must be a non-empty string")
  }
  if len(args[3]) <= 0 {
    return shim.Error("4th argument must be a non-empty string")
  }
  if len(args[4]) <= 0 {
    return shim.Error("5th argument must be a non-empty string")
  }
  if len(args[5]) <= 0 {
    return shim.Error("6th argument must be a non-empty string")
  }

  eventName := args[0]
  debtorName := args[1]
  creditorName := args[2]
  itemName := args[3]
  itemPrice := args[4]
  itemQuality := args[5]

  // ==== Build the commitment json string manually if you don't want to use struct marshalling ====
  commitmentJSONasString := `
    {
      "docType":"Event",
      "debtorName": "` + debtorName + `",
      "creditorName": "` + creditorName + `",
      "event": "` + eventName + `",
      "item": "` + itemName + `",
      "price": "` + itemPrice + `",
      "quality": "` + itemQuality + `"
    }`
  commitmentJSONasBytes := []byte(commitmentJSONasString)

  // === Save commitment to state ===
  ID, err := genUUIDv4()
  err = stub.PutState(eventName + ID.String(), commitmentJSONasBytes)
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Index the commitment to enable color-based range queries, e.g. return all blue commitments ====
  //  An 'index' is a normal key/value entry in state.
  //  The key is a composite key, with the elements that you want to range query on listed first.
  //  In our case, the composite key is based on indexName~name.
  //  This will enable very efficient state range queries based on composite keys matching indexName~color~*
  indexName := "event"
  ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{eventName})
  if err != nil {
    return shim.Error(err.Error())
  }

  //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data.
  //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
  value := []byte{0x00}
  stub.PutState(ownerNameIndexKey, value)

  // ==== Data saved and indexed. Return success ====
  fmt.Println("- end init commitment data")

  // Notify listeners that an event "eventInvoke" have been executed (check line 24 in the file invoke.go)
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(nil)
}

// ========================================================
// readCommitment - read a commitment from chaincode state
// ========================================================
func (t *HeroesServiceChaincode) readCommitment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var name, jsonResp string
  var err error

  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting name of the commitment to query")
  }

  // Get the commitment from chaincode state
  name = args[0]
  valAsbytes, err := stub.GetState(name)
  if err != nil {
    jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
    return shim.Error(jsonResp)
  } else if valAsbytes == nil {
    jsonResp = "{\"Error\":\"Commitment does not exist: " + name + "\"}"
    return shim.Error(jsonResp)
  }

  // Notify listeners that an event "eventInvoke" have been executed (check line 19 in the file invoke.go)
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(valAsbytes)
}

// ===== Example: Ad hoc rich query ========================================================
// richQuery uses a query string to perform a query for commitments.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *HeroesServiceChaincode) richQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {

  //   0
  // "queryString"
  if len(args) < 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
  }

  queryString := args[0]
  queryResults, err := getQueryResultForQueryString(stub, queryString)
  if err != nil {
    return shim.Error(err.Error())
  }
  return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

  fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

  resultsIterator, err := stub.GetQueryResult(queryString)
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  buffer, err := constructQueryResponseFromIterator(resultsIterator)
  if err != nil {
    return nil, err
  }

  fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

  return buffer.Bytes(), nil
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
  // buffer is a JSON array containing QueryResults
  var buffer bytes.Buffer
  buffer.WriteString("[")

  bArrayMemberAlreadyWritten := false
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    // Add a comma before array members, suppress it for the first array member
    if bArrayMemberAlreadyWritten == true {
      buffer.WriteString(",")
    }
    buffer.WriteString("{\"Key\":")
    buffer.WriteString("\"")
    buffer.WriteString(queryResponse.Key)
    buffer.WriteString("\"")

    buffer.WriteString(", \"Record\":")
    // Record is a JSON object, so we write as-is
    buffer.WriteString(string(queryResponse.Value))
    buffer.WriteString("}")
    bArrayMemberAlreadyWritten = true
  }
  buffer.WriteString("]")

  return &buffer, nil
}

// invoke
// Every functions that read and write in the ledger will be here
func (t *HeroesServiceChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  fmt.Println("########### HeroesServiceChaincode invoke ###########")

  if len(args) < 2 {
    return shim.Error("The number of arguments is insufficient.")
  }

  // Check if the ledger key is "hello" and process if it is the case. Otherwise it returns an error.
  if args[0] == "hello" && len(args) == 2 {

    // Write the new value in the ledger
    err := stub.PutState("hello", []byte(args[1]))
    if err != nil {
      return shim.Error("Failed to update state of hello")
    }

    // Notify listeners that an event "eventInvoke" have been executed (check line 19 in the file invoke.go)
    err = stub.SetEvent("eventInvoke", []byte{})
    if err != nil {
      return shim.Error(err.Error())
    }

    // Return this value in response
    return shim.Success(nil)
  }

  // If the arguments given don’t match any function, we return an error
  return shim.Error("Unknown invoke action, check the second argument.")
}

// query
// Every readonly functions in the ledger will be here
func (t *HeroesServiceChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  fmt.Println("########### HeroesServiceChaincode query ###########")

  // Check whether the number of arguments is sufficient
  if len(args) < 1 {
    return shim.Error("The number of arguments is insufficient.")
  }

  // Like the Invoke function, we manage multiple type of query requests with the second argument.
  // We also have only one possible argument: hello
  if args[0] == "hello" {

    // Get the state of the value matching the key hello in the ledger
    state, err := stub.GetState("hello")
    if err != nil {
      return shim.Error("Failed to get state of hello")
    }

    // Return this value in response
    return shim.Success(state)
  }

  // If the arguments given don’t match any function, we return an error
  return shim.Error("Unknown query action, check the second argument.")
}

func genUUIDv4() (uuid.UUID, error) {
  return uuid.NewV4()
}

func main() {
  // Start the chaincode and make it ready for futures requests
  err := shim.Start(new(HeroesServiceChaincode))
  if err != nil {
    fmt.Printf("Error starting Heroes Service chaincode: %s", err)
  }
}
