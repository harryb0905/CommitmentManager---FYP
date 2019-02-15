package main

import (
  "fmt"
  "encoding/json"
  "bytes"
  "encoding/binary"
  "log"
  // "github.com/davecgh/go-spew/spew"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  pb "github.com/hyperledger/fabric/protos/peer"
  q "github.com/scc300/scc300-network/chaincode/commitments/quark"
)

const (
  GreenTick = "\033[92m" + "\u2713" + "\033[0m"
  GetEventQuery = "{\"selector\":{\"docType\":\"%s\"}}"
  TimeFormat = "Mon Jan _2 15:04:05 2006"
)

// SCC300NetworkChaincode implementation of Chaincode
type SCC300NetworkChaincode struct {
}

type Spec struct {
  ObjectType  string `json:"docType"`     // docType is used to distinguish the various types of objects in state database
  Name        string `json:"name"`        // Spec name - the field tags are needed to keep case from bouncing around
  Source      string `json:"source"`      // String to store spec source code (quark)
}

type Commitment struct {
  ComID    string
  States []ComState
}

type ComState struct {
  Name  string
  Data  map[string]interface{}
}

type QueryResponse struct {
  Key     string
  Record  map[string]interface{}
}

type CommitmentMeta struct {
  Name     string  `json:"name"`
  Source   string  `json:"source"`
  Summary  string  `json:"summary"` // may not need...
}

// ============================================================
// Init - This function is called only one when the chaincode is instantiated.
// Goal is to prepare the ledger to handle future requests.
// ============================================================
func (t *SCC300NetworkChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

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
func (t *SCC300NetworkChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

  // ==== Get the function and arguments from the request ====
  function, args := stub.GetFunctionAndParameters()

  // ==== Check whether the number of arguments is sufficient ====
  if len(args) < 1 {
    return shim.Error("The number of arguments is insufficient.")
  }

  // ==== Handle different functions ====
  if function == "initSpec" {
    return t.initSpec(stub, args)
  } else if function == "getSpec" {
    return t.getSpec(stub, args)
  } else if function == "getCreatedCommitments" {
    return t.getCreatedCommitments(stub, args)
  } else if function == "initCommitmentData" {
    return t.initCommitmentData(stub, args)
  } else if function == "richQuery" {
    return t.richQuery(stub, args)
  }

  // ==== If the arguments given don’t match any function, we return an error ====
  return shim.Error("Unknown action, check the first argument")
}

// =======================================================================
// initSpec - create a new spec, store into chaincode state.
// The argument list consists of the spec name and the spec source code.
// =======================================================================
func (t *SCC300NetworkChaincode) initSpec(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var err error

  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting <spec_source>")
  }

  // ==== Input sanitation ====
  fmt.Println("- start init spec")
  if len(args[0]) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }

  // ==== Get spec source from arg list ==== //
  source := args[0]

  // ==== Compile the specification on the chaincode ==== //
  // This obtains meta info about the spec ready to initialise on CouchDB ==== //
  spec, err := compileSpec(source)
  specName := spec.Constraint.Name

  // ==== Check if spec already exists ====
  specAsBytes, err := stub.GetState(specName)
  if err != nil {
    return shim.Error("Failed to get spec: " + err.Error())
  } else if specAsBytes != nil {
    fmt.Println("This spec already exists: " + specName)
    return shim.Error("This spec already exists: " + specName)
  }

  // ==== Create spec object and marshal to JSON ====
  objectType := "spec"
  specRes := &Spec{objectType, specName, source}
  specJSONasBytes, err := json.Marshal(specRes)
  if err != nil {
    return shim.Error(err.Error())
  }

  // ==== Save spec to state ====
  err = stub.PutState(specName, specJSONasBytes)
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Index the spec to enable range-based queries, e.g. return all SellItem commitments ====
  // indexName := "owner~name"
  indexName := "name"
  ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{specRes.Name})
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the commitment. ====
  //  ==== Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value. ====
  value := []byte{0x00}
  stub.PutState(ownerNameIndexKey, value)

  // ==== Spec saved and indexed. Return success ====
  fmt.Println("- end init spec")

  // ==== Notify listeners that an event "eventInvoke" have been executed (see invoke.go) ====
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(nil)
}

// ========================================================
// getSpec - read a specification from chaincode state.
// ========================================================
func (t *SCC300NetworkChaincode) getSpec(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var name, jsonResp string
  var err error

  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting name of the spec to query")
  }

  // ==== Get the spec from chaincode state ====
  name = args[0]
  valAsbytes, err := stub.GetState(name)
  if err != nil {
    jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
    return shim.Error(jsonResp)
  } else if valAsbytes == nil {
    jsonResp = "{\"Error\":\"Spec does not exist: " + name + "\"}"
    return shim.Error(jsonResp)
  }

  // ==== Notify listeners that an event "eventInvoke" have been executed (check line 19 in the file invoke.go) ====
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(valAsbytes)
}

// =========================== GET CREATED COMMITMENTS ===================================
//  Obtains all created commitments based on a given commitment/spec name.
//  A commitment is created if it exists on the blockchain CouchDB database.
// =======================================================================================
func (t *SCC300NetworkChaincode) getCreatedCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  comName := args[0]
  commitments := []Commitment{}

  // Obtain spec from CouchDB based on the comName
  response := t.getSpec(stub, []string{comName})
  res := string(response.Payload)

  // Obtain spec from CouchDB based on the comName
  com := CommitmentMeta{}

  // Unmarshal JSON into structure and obtain source code
  json.Unmarshal([]byte(res), &com)

  // ==== Compile specification source to obtain struct ==== //
  specSource := com.Source
  spec, _ := compileSpec(specSource)

  // Format query and perform query to get created commitment results
  query := fmt.Sprintf(GetEventQuery, spec.CreateEvent.Name)
  queryArgs := []string{query}
  queryRes := t.richQuery(stub, queryArgs)
  queryResponsesPayload := queryRes.Payload

  // Unmarshal JSON response from query
  responses := []QueryResponse{}
  json.Unmarshal([]byte(queryResponsesPayload), &responses)

  // Create commitments from responses with data per commitment state
  for _, elem := range responses {
   commitments = append(commitments,
     Commitment{
       ComID: elem.Record["comID"].(string),
       States: []ComState {
         ComState{
           Name: "Created",
           Data: elem.Record,
         },
         ComState{Name: "Detached", Data: nil,},
         ComState{Name: "Discharged", Data: nil,},
       },
     },
   )
  }
  // return commitments, com, err;

  buf := new(bytes.Buffer)
  b, err := json.Marshal(commitments)
  if err != nil {
    return shim.Error(err.Error())
  }

  err = binary.Write(buf, binary.BigEndian, &b)
  if err != nil {
    fmt.Println("binary.Write failed:", err)
  }

  fmt.Println(buf.Bytes())
  fmt.Println(":::CREATED COMMITMENTS:::", commitments)
  return shim.Success(buf.Bytes())
  // return nil, com, fmt.Errorf("Couldn't get %s created commitments", comName)
}

// ======================================================================
// initCommitmentData - adds commitment data to blockchain to be queried.
// Accepts an array of JSON object strings and adds to CouchDB.
// ======================================================================
func (t *SCC300NetworkChaincode) initCommitmentData(stub shim.ChaincodeStubInterface, args []string) pb.Response {

  // === Add slice data to database ===
  for _, commitmentDataJSON := range args {
    fmt.Println("STR:", commitmentDataJSON)
    commitmentDataJSONBytes := []byte(commitmentDataJSON)

    // === Obtain event name from current JSON string ===
    var jsonMap map[string]string
    json.Unmarshal([]byte(commitmentDataJSON), &jsonMap)
    eventName := string(jsonMap["docType"])
    comID := string(jsonMap["comID"])

    // === Save commitment to state creating a new instance with an id ===
    err := stub.PutState(eventName + comID, commitmentDataJSONBytes)
    if err != nil {
      return shim.Error(err.Error())
    }

    //  ==== Index the commitment to enable event name-based range queries ====
    indexName := "event"
    ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{eventName})
    if err != nil {
      return shim.Error(err.Error())
    }

    //  === Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data. ===
    //  === Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value. ===
    value := []byte{0x00}
    stub.PutState(ownerNameIndexKey, value)

    // === Data saved and indexed. Return success ===
    fmt.Println("- end init commitment data")

    // === Notify listeners that an event "eventInvoke" have been executed (check line 24 in the file invoke.go) ===
    err = stub.SetEvent("eventInvoke", []byte{})
    if err != nil {
      return shim.Error(err.Error())
    }
  }

  return shim.Success(nil)
}

// =========================================================================================
// richQuery - uses a query string to perform a query for commitments.
//
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// Only available on state databases that support rich query (e.g. CouchDB).
// The first argument in the args list is the query string.
// =========================================================================================
func (t *SCC300NetworkChaincode) richQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {

  // ==== Input sanitation =====
  if len(args) < 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
  }

  // ==== Obtain query results ====
  queryString := args[0]
  queryResults, err := getQueryResultForQueryString(stub, queryString)
  if err != nil {
    return shim.Error(err.Error())
  }
  return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString - executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

  // ==== Obtain query result ====
  resultsIterator, err := stub.GetQueryResult(queryString)
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  // ==== Construct query response ====
  buffer, err := constructQueryResponseFromIterator(resultsIterator)
  if err != nil {
    return nil, err
  }

  return buffer.Bytes(), nil
}

// ===========================================================================================
// constructQueryResponseFromIterator - constructs a JSON array containing query results from
// a given result iterator.
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {

  // ==== Buffer is a JSON array containing QueryResults ====
  var buffer bytes.Buffer
  buffer.WriteString("[")

  bArrayMemberAlreadyWritten := false
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    // ==== Add a comma before array members, suppress it for the first array member ====
    if bArrayMemberAlreadyWritten == true {
      buffer.WriteString(",")
    }
    buffer.WriteString("{\"Key\":")
    buffer.WriteString("\"")
    buffer.WriteString(queryResponse.Key)
    buffer.WriteString("\"")

    buffer.WriteString(", \"Record\":")
    // ==== Record is a JSON object, so we write as-is ====
    buffer.WriteString(string(queryResponse.Value))
    buffer.WriteString("}")
    bArrayMemberAlreadyWritten = true
  }
  buffer.WriteString("]")

  return &buffer, nil
}

// Function to obtain specification info based on file input data
// A list of args for initialisation on the blockchain is created
// func getSpecInfo(specSource string) (meta CommitmentMeta, spec Spec, err error) {
//   spec, err := q.Parse(source)
//   if (err != nil) {
//     log.Fatal("\nSyntax Error:\n", err, "\n")
//   } else {
//     fmt.Printf("\n%s spec compiled successfully %s \n\n", spec.Constraint.Name, GreenTick)
//   }

//   // Obtain spec from CouchDB based on the comName
//   com := CommitmentMeta{}
//   response, er := t.getSpec(comName)
//   if (er != nil) {
//     return com, nil, er
//   }

//   // Unmarshal JSON into structure and obtain source code
//   json.Unmarshal([]byte(response), &com)

//   // Compile specification (using custom built compiler) to obtain events
//   spec, _ = q.Parse(com.Source)
//   return com, spec, nil;

//   // Create list of args to initialise a new spec
//   // specArgs := []string{spec.Constraint.Name, source}
//   // return specArgs
// }

func compileSpec(source string) (res *q.Spec, err error) {
  spec, err := q.Parse(source)
  if (err != nil) {
    log.Fatal("\nSyntax Error:\n", err, "\n")
  } else {
    fmt.Printf("\n%s spec compiled successfully %s \n\n", spec.Constraint.Name, GreenTick)
  }
  return spec, err
}

// ==================================================================
// main - start the chaincode and make it ready for future requests.
// ==================================================================
func main() {
  err := shim.Start(new(SCC300NetworkChaincode))
  if err != nil {
    fmt.Printf("Error starting SCC300Network chaincode: %s", err)
  }
}
