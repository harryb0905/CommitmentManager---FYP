package main

import (
  "fmt"
  "time"
  "math"
  "strconv"
  "bytes"
  "encoding/binary"
  "encoding/json"
  "log"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  pb "github.com/hyperledger/fabric/protos/peer"
  q "github.com/scc300/scc300-network/chaincode/quark"
)

const (
  GetEventQuery = "{\"selector\":{\"docType\":\"%s\"}}"  // Obtains all event data based on docType

  GreenTick = "\033[92m" + "\u2713" + "\033[0m"
  TimeFormat = "Mon Jan _2 15:04:05 2006"
)

// SCC300NetworkChaincode implementation of Chaincode
type SCC300NetworkChaincode struct {
}

type Spec struct {
  ObjectType  string `json:"docType"`  // docType - used to distinguish the various types of objects in state database
  Name        string `json:"name"`     // Spec name - the name of the specification
  Source      string `json:"source"`   // Source - string to store spec source code (.quark file)
}

type Commitment struct {
  ComID    string     // ComID - stores this commitment ID (each commitment is unique)
  States []ComState   // States - slice of commitment states 
}

type ComState struct {
  Name  string                    // Name - name of this particular commitment state (i.e. created, detached, discharged, expired, violated)
  Data  map[string]interface{}    // Data - map of data associated with this state
}

type QueryResponse struct {
  Key     string                  // Key - the key for this query response
  Record  map[string]interface{}  // Record - the record associated with this key for this query response
}

// =============================================================================
// Init - This function is called only once when the chaincode is instantiated.
// Goal is to prepare the ledger to handle future requests.
// =============================================================================
func (t *SCC300NetworkChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

  // ==== Get the function and arguments from the request ==== //
  function, _ := stub.GetFunctionAndParameters()

  // ==== Check if the request is the init function ==== //
  if function != "init" {
    return shim.Error("Unknown function call")
  }

  // ==== Return a successful message ==== //
  return shim.Success(nil)
}

// ============================================================
// Invoke - All future invoke requests will arrive here
// ============================================================
func (t *SCC300NetworkChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

  // ==== Get the function and arguments from the request ==== //
  function, args := stub.GetFunctionAndParameters()

  // ==== Check whether the number of arguments is sufficient ==== //
  if len(args) < 1 {
    return shim.Error("The number of arguments is insufficient.")
  }

  // ==== Handle different functions ==== //
  if function == "initSpec" {
    return t.initSpec(stub, args)
  } else if function == "getSpec" {
    return t.getSpec(stub, args)
  } else if function == "initCommitmentData" {
    return t.initCommitmentData(stub, args)
  } else if function == "richQuery" {
    return t.richQuery(stub, args)
  } else if function == "getCreatedCommitments" {
    return t.getCreatedCommitments(stub, args)
  } else if function == "getDetachedCommitments" {
    return t.getDetachedCommitments(stub, args)
  } else if function == "getExpiredCommitments" {
    return t.getExpiredCommitments(stub, args)
  } else if function == "getDischargedCommitments" {
    return t.getDischargedCommitments(stub, args)
  } else if function == "getViolatedCommitments" {
    return t.getViolatedCommitments(stub, args)
  } 

  // ==== If the arguments given don’t match any function, we return an error ==== //
  return shim.Error("Unknown action, check the first argument")
}

// =======================================================================
// initSpec - create a new spec, store into chaincode state.
// The argument list consists of the spec name and the spec source code.
// =======================================================================
func (t *SCC300NetworkChaincode) initSpec(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  var err error

  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting <specSource>")
  }

  // ==== Input sanitation ==== //
  fmt.Println("- start init spec")
  if len(args[0]) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }

  // ==== Get spec source from arg list ==== //
  source := args[0]

  // ==== Compile the specification on the chaincode ==== //
  // ==== This obtains meta info about the spec ready to initialise on CouchDB ==== //
  spec, err := compileSpec(source)
  specName := spec.Constraint.Name

  // ==== Check if spec already exists ==== //
  specAsBytes, err := stub.GetState(specName)
  if err != nil {
    return shim.Error("Failed to get spec: " + err.Error())
  } else if specAsBytes != nil {
    fmt.Println("This spec already exists: " + specName)
    return shim.Error("This spec already exists: " + specName)
  }

  // ==== Create spec object and marshal to JSON ==== //
  objectType := "spec"
  specRes := &Spec{objectType, specName, source}
  specJSONasBytes, err := json.Marshal(specRes)
  if err != nil {
    return shim.Error(err.Error())
  }

  // ==== Save spec to state ==== //
  err = stub.PutState(specName, specJSONasBytes)
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Index the spec to enable range-based queries, e.g. return all SellItem commitments ==== //
  indexName := "name"
  ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{specRes.Name})
  if err != nil {
    return shim.Error(err.Error())
  }

  //  ==== Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the commitment ==== //
  //  ==== Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value ==== //
  value := []byte{0x00}
  stub.PutState(ownerNameIndexKey, value)

  // ==== Spec saved and indexed. Return success ==== //
  fmt.Println("- end init spec")

  // ==== Notify listeners that an event "eventInvoke" have been executed (see invoke.go) ==== //
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

  // ==== Get the spec from chaincode state ==== //
  name = args[0]
  valAsbytes, err := stub.GetState(name)
  if err != nil {
    jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
    return shim.Error(jsonResp)
  } else if valAsbytes == nil {
    jsonResp = "{\"Error\":\"Spec does not exist: " + name + "\"}"
    return shim.Error(jsonResp)
  }

  // ==== Notify listeners that an event "eventInvoke" have been executed (check line 19 in the file invoke.go) ==== //
  err = stub.SetEvent("eventInvoke", []byte{})
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(valAsbytes)
}

// ======================================================================
// initCommitmentData - adds commitment data to blockchain to be queried.
// Accepts an array of JSON object strings and adds to CouchDB.
// ======================================================================
func (t *SCC300NetworkChaincode) initCommitmentData(stub shim.ChaincodeStubInterface, args []string) pb.Response {

  // ==== Add slice data to database ==== //
  for _, commitmentDataJSON := range args {
    commitmentDataJSONBytes := []byte(commitmentDataJSON)

    // ==== Obtain event name from current JSON string ==== //
    var jsonMap map[string]string
    json.Unmarshal([]byte(commitmentDataJSON), &jsonMap)
    eventName := string(jsonMap["docType"])
    comID := string(jsonMap["comID"])

    // ==== Save commitment to state creating a new instance with an id ==== //
    err := stub.PutState(eventName + comID, commitmentDataJSONBytes)
    if err != nil {
      return shim.Error(err.Error())
    }

    //  ==== Index the commitment to enable event name-based range queries ==== //
    // indexName := "event"
    // ownerNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{eventName})
    // if err != nil {
    //   return shim.Error(err.Error())
    // }

    //  ==== Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the data. ==== //
    //  ==== Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value. ==== //
    // value := []byte{0x00}
    // stub.PutState(ownerNameIndexKey, value)

    // ==== Data saved and indexed. Return success ==== //
    fmt.Println("- end init commitment data")

    // ==== Notify listeners that an event "eventInvoke" have been executed (check line 24 in the file invoke.go) ==== //
    err = stub.SetEvent("eventInvoke", []byte{})
    if err != nil {
      return shim.Error(err.Error())
    }
  }

  return shim.Success(nil)
}

// =============================== COMMITMENT API METHODS ======================================== //
//
//  getCreatedCommitments(stub, args): obtains all created commitments by commitment name.
//    - stub: required chaincode interface
//    - args: slice of strings (args[0]: commitment name)
//  getDetachedCommitments(stub, args): obtains all detached commitments by commitment name.
//    - stub: required chaincode interface
//    - args: slice of strings (args[0]: commitment name, args[1]: false)
//  getDischargedCommitments(stub, args): obtains all discharged commitments by commitment name.
//    - stub: required chaincode interface
//    - args: slice of strings (args[0]: commitment name, args[1]: false)
//  getExpiredCommitments(stub, args): obtains all expired commitments by commitment name.
//    - stub: required chaincode interface
//    - args: slice of strings (args[0]: commitment name, args[1]: true (boolean flag - true to
//            get expired, false to get detached - this prevents repetition of logic))
//  getViolatedCommitments(stub, args): obtains all violated commitments by commitment name.
//    - stub: required chaincode interface
//    - args: slice of strings (args[0]: commitment name, args[1]: true (boolean flag - true to
//            get violated, false to get discharged - this prevents repetition of logic))
//
// =============================================================================================== //

// =========================== GET CREATED COMMITMENTS ========================
//  Obtains all created commitments based on a given commitment name.
//  A commitment is created if it exists on the blockchain CouchDB database.
// ============================================================================
func (t *SCC300NetworkChaincode) getCreatedCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  comName := args[0]
  commitments := []Commitment{}

  // ==== Obtain spec from CouchDB based on the comName ==== //
  response := t.getSpec(stub, []string{comName})
  res := string(response.Payload)

  // ==== Unmarshal JSON into structure and obtain source code ==== //
  com := Spec{}
  json.Unmarshal([]byte(res), &com)

  // ==== Compile specification source to obtain go struct ==== //
  spec, _ := compileSpec(com.Source)

  // ==== Format and perform query to get created commitment results ==== //
  query := fmt.Sprintf(GetEventQuery, spec.CreateEvent.Name)
  queryArgs := []string{query}
  queryRes := t.richQuery(stub, queryArgs)
  queryResponsesPayload := queryRes.Payload

  // ==== Unmarshal JSON response from query ==== //
  responses := []QueryResponse{}
  json.Unmarshal([]byte(queryResponsesPayload), &responses)

  // ==== Create commitments from responses with data per commitment state ==== //
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
  // ==== Convert commitments to bytes to send to requester ==== //
  commitmentsBytes, _ := commitmentsToBytes(commitments)
  return shim.Success(commitmentsBytes)
}

// =========================== GET DETACHED COMMITMENTS ======================================
//  Obtains all detached commitments based on a given commitment/spec name.
//  A commitment is detached if the created event exists on the blockchain CouchDB database
//  and the detached event has occured within the specified deadline.
//  If the commitment isn't detached and the deadline has exceeded, the commitment expires.
// ===========================================================================================
func (t *SCC300NetworkChaincode) getDetachedCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  // ==== Create commitments from responses with data per commitment state ==== //
  commitments := []Commitment{}

  // ==== Extract args ==== //
  if len(args) != 2 {
    return shim.Error("Incorrect number of arguments. Expecting [<comName>, <wantExpired>]")
  }

  comName := args[0]
  wantExpired, _ := strconv.ParseBool(args[1])

  // ==== Input sanitation ====
  if len(comName) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }

  // ==== Obtain spec from CouchDB based on the comName ==== //
  response := t.getSpec(stub, []string{comName})
  res := string(response.Payload)

  // ==== Unmarshal JSON into structure and obtain source code ==== //
  com := Spec{}
  json.Unmarshal([]byte(res), &com)

  // ==== Compile specification source to obtain struct ==== //
  specSource := com.Source
  spec, _ := compileSpec(specSource)

  // ==== Check if this commitment has been created (can't be detached if not already created) ==== //
  createdResponse := t.getCreatedCommitments(stub, []string{comName})
  createdComs := []Commitment{}
  json.Unmarshal([]byte(createdResponse.Payload), &createdComs)

  // ==== Extract the deadline value from the spec source (e.g. deadline=5) ==== //
  deadline := getDeadline(spec.DetachEvent.Args)

  // ==== Format and perform query to get detached commitment results ==== //
  query := fmt.Sprintf(GetEventQuery, spec.DetachEvent.Name)
  queryArgs := []string{query}
  queryRes := t.richQuery(stub, queryArgs)
  queryResponsesPayload := queryRes.Payload

  // ==== Unmarshal JSON response from query ==== //
  responses := []QueryResponse{}
  json.Unmarshal([]byte(queryResponsesPayload), &responses)
  hasDetachedEvent := false

  // ==== Date checks with deadlines on detached results ==== //
  // ==== If detach event record exists and that timestamp is within a period of x days or less from offer being created, this commitment is detached ==== //
  for _, createdCom := range createdComs {
    for _, comRes := range responses {
    // ==== Get created commitment that corresponds to this detached commitment ==== //
      if (createdCom.ComID == comRes.Record["comID"].(string)) {
        hasDetachedEvent = true
        // ==== Extract date for checking deadline ==== //
        createdDateStr := createdCom.States[0].Data["date"].(string)
        detachedDateStr := comRes.Record["date"].(string)

        // ==== If detached event date is within specified deadline, include in results ==== //
        withinDeadline := isDateWithinDeadline(createdDateStr, detachedDateStr, deadline)
        if ((withinDeadline && !wantExpired) || (!withinDeadline && wantExpired)) {
          commitments = append(commitments,
            Commitment{
              ComID: createdCom.ComID,
              States: []ComState {
                ComState{
                  Name: "Created",
                  Data: createdCom.States[0].Data,
                 },
                ComState{
                  Name: "Detached",
                  Data: comRes.Record,
                },
                ComState{Name: "Discharged", Data: nil,},
              },
            },
          )
        }
        break
      }
    }

    // ==== Edge case where create event exists but detach event doesn't ==== //
    // ==== (i.e. use todays date to determine if it should be detached ==== //
    if (!hasDetachedEvent) {

      // ==== Extract dates for checking deadline ==== //
      createdDateStr := createdCom.States[0].Data["date"].(string)
      todayStr := time.Now().String()

      // ==== If detach event date is within specified deadline, include in results ==== //
      withinDeadline := isDateWithinDeadline(createdDateStr, todayStr, deadline)
      if (!withinDeadline && wantExpired) {
        commitments = append(commitments,
          Commitment{
            ComID: createdCom.ComID,
            States: []ComState {
              ComState{
                Name: "Created",
                Data: createdCom.States[0].Data,
              },
              ComState{
                Name: "Detached",
                Data: nil,
              },
              ComState{Name: "Discharged", Data: nil,},
            },
          },
        )
      }
    }
    hasDetachedEvent = false
  }
  // ==== Convert commitments to bytes to send to requester ==== //
  commitmentsBytes, _ := commitmentsToBytes(commitments)
  return shim.Success(commitmentsBytes)
}

// =========================== GET DISCHARGED COMMITMENTS =======================
//  Obtains all discharged commitments based on a given commitment/spec name.
// ==============================================================================
func (t *SCC300NetworkChaincode) getDischargedCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  // ==== Create commitments from responses with data per commitment state ==== //
  commitments := []Commitment{}

  // ==== Extract args ==== //
  if len(args) != 2 {
    return shim.Error("Incorrect number of arguments. Expecting [<comName>, <wantViolated>]")
  }

  comName := args[0]
  wantViolated, _ := strconv.ParseBool(args[1])

  // ==== Input sanitation ====
  if len(comName) <= 0 {
    return shim.Error("1st argument must be a non-empty string")
  }

  // ==== Obtain spec from CouchDB based on the comName ==== //
  response := t.getSpec(stub, []string{comName})
  res := string(response.Payload)

  // ==== Unmarshal JSON into structure and obtain source code ==== //
  com := Spec{}
  json.Unmarshal([]byte(res), &com)

  // ==== Compile specification source to obtain struct ==== //
  specSource := com.Source
  spec, _ := compileSpec(specSource)

  // ==== Check if this commitment has been detached (can't be discharged if not already detached) ==== //
  // ==== Extra false arg is required to only get the detached commitments (i.e. don't want expired) ==== //
  detachedResponse := t.getDetachedCommitments(stub, []string{comName, "false"})
  detachedComs := []Commitment{}
  json.Unmarshal([]byte(detachedResponse.Payload), &detachedComs)

  // ==== Extract the deadline value from the spec source (e.g. deadline=5) ==== //
  deadline := getDeadline(spec.DischargeEvent.Args)

  // ==== Format and perform query to get discharged commitment results ==== //
  query := fmt.Sprintf(GetEventQuery, spec.DischargeEvent.Name)
  queryArgs := []string{query}
  queryRes := t.richQuery(stub, queryArgs)
  queryResponsesPayload := queryRes.Payload

  // ==== Unmarshal JSON response from query ==== //
  responses := []QueryResponse{}
  json.Unmarshal([]byte(queryResponsesPayload), &responses)
  hasDischargedEvent := false

  // ==== Date checks with deadlines on discharged results ==== //
  // ==== If discharge event record exists and that timestamp is within a period of x days or less from offer being detached, this commitment is discharged ==== //
  for _, detachedCom := range detachedComs {
    for _, comRes := range responses {
      // ==== Get detachec commitment that corresponds to this detached commitment ==== //
      if (detachedCom.ComID == comRes.Record["comID"].(string)) {
        hasDischargedEvent = true
        // ==== Extract date for checking deadline ==== //
        createdDateStr := detachedCom.States[0].Data["date"].(string)
        dischargedDateStr := comRes.Record["date"].(string)

        // ==== If discharge event date is within specified deadline, include in results ==== //
        withinDeadline := isDateWithinDeadline(createdDateStr, dischargedDateStr, deadline)
        if ((withinDeadline && !wantViolated) || (!withinDeadline && wantViolated)) {
          commitments = append(commitments,
            Commitment{
              ComID: comRes.Record["comID"].(string),
              States: []ComState {
                ComState{
                  Name: "Created",
                  Data: detachedCom.States[0].Data,
                },
                ComState{
                  Name: "Detached",
                  Data: detachedCom.States[1].Data,
                },
                ComState{
                  Name: "Discharged",
                  Data: comRes.Record,
                },
              },
            },
          )
        }
        break
      }
    }

    // ==== Edge case where detach event exists but discharge event doesn't ==== //
    // ==== (i.e. use todays date to determine if it should be discharged ==== //
    if (!hasDischargedEvent) {

      // ==== Extract dates for checking deadline ==== //
      detachedDateStr := detachedCom.States[1].Data["date"].(string)
      todayStr := time.Now().String()

      // ==== If discharge event date is within specified deadline, include in results ==== //
      withinDeadline := isDateWithinDeadline(detachedDateStr, todayStr, deadline)
      if (!withinDeadline && wantViolated) {
        commitments = append(commitments,
          Commitment{
            ComID: detachedCom.ComID,
            States: []ComState {
              ComState{
                Name: "Created",
                Data: detachedCom.States[0].Data,
              },
              ComState{
                Name: "Detached",
                Data: detachedCom.States[1].Data,
              },
              ComState{
                Name: "Discharged",
                Data: nil,
              },
            },
          },
        )
      }
    }
    hasDischargedEvent = false
  }
  // ==== Convert commitments to bytes to send to requester ==== //
  commitmentsBytes, _ := commitmentsToBytes(commitments)
  return shim.Success(commitmentsBytes)
}

// =========================== GET EXPIRED COMMITMENTS ======================
//  Obtains all expired commitments based on a given commitment/spec name.
//  This method simply calls getDetachedCommitments() with an extra
//  boolean flag of 'true' to obtain all the failed detached commitments.
// ==========================================================================
func (t *SCC300NetworkChaincode) getExpiredCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  return t.getDetachedCommitments(stub, args)
}

// =========================== GET VIOLATED COMMITMENTS ======================
//  Obtains all violated commitments based on a given commitment/spec name.
//  This method simply calls getDischargedCommitments() with an extra
//  boolean flag of 'true' to obtain all the failed discharged commitments.
// ===========================================================================
func (t *SCC300NetworkChaincode) getViolatedCommitments(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  return t.getDischargedCommitments(stub, args)
}

// ===============================================================================
// richQuery - uses a query string to perform a query for commitments.
//
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// Only available on state databases that support rich query (e.g. CouchDB).
// The first argument in the args list is the query string.
// ===============================================================================
func (t *SCC300NetworkChaincode) richQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {

  // ==== Input sanitation ===== //
  if len(args) < 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
  }

  // ==== Obtain query results ==== //
  queryString := args[0]
  queryResults, err := getQueryResultForQueryString(stub, queryString)
  if err != nil {
    return shim.Error(err.Error())
  }
  return shim.Success(queryResults)
}

// =================================================================================
// getQueryResultForQueryString - executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

  // ==== Obtain query result ==== //
  resultsIterator, err := stub.GetQueryResult(queryString)
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  // ==== Construct query response ==== //
  buffer, err := constructQueryResponseFromIterator(resultsIterator)
  if err != nil {
    return nil, err
  }

  return buffer.Bytes(), nil
}

// ============================================================================================
// constructQueryResponseFromIterator - constructs a JSON array containing query results from
// a given result iterator.
// ============================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {

  // ==== Buffer is a JSON array containing QueryResults ==== //
  var buffer bytes.Buffer
  buffer.WriteString("[")

  bArrayMemberAlreadyWritten := false
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    // ==== Add a comma before array members, suppress it for the first array member ==== //
    if bArrayMemberAlreadyWritten == true {
      buffer.WriteString(",")
    }
    buffer.WriteString("{\"Key\":")
    buffer.WriteString("\"")
    buffer.WriteString(queryResponse.Key)
    buffer.WriteString("\"")

    buffer.WriteString(", \"Record\":")
    // ==== Record is a JSON object, so we write as-is ==== //
    buffer.WriteString(string(queryResponse.Value))
    buffer.WriteString("}")
    bArrayMemberAlreadyWritten = true
  }
  buffer.WriteString("]")

  return &buffer, nil
}

// ======================================================================
// compileSpec - compiles a specification source code into a go struct.
// ======================================================================
func compileSpec(source string) (res *q.Spec, err error) {
  spec, err := q.Parse(source)
  if (err != nil) {
    log.Fatal("\nSyntax Error:\n", err, "\n")
  } else {
    fmt.Printf("\n%s spec compiled successfully %s \n\n", spec.Constraint.Name, GreenTick)
  }
  return spec, err
}

// ======================================================================================
// isDateWithinDeadline - perform Go time arithmetic on dates with specified deadline.
// (e.g. deadline=5 means payment must occur within 5 days of the offer being created)
// ======================================================================================
func isDateWithinDeadline(date1 string, date2 string, deadline float64) (within bool) {
  createEventDate, _ := time.Parse(TimeFormat, date1)
  detachEventDate, _ := time.Parse(TimeFormat, date2)
  daysDiff := float64(detachEventDate.Sub(createEventDate).Hours() / 24)

  if (math.Abs(daysDiff) >= deadline) {
    return false
  } else {
    return true
  }
}

// ======================================================================
// getDeadline - obtains the deadline value from the list of arguments.
// ======================================================================
func getDeadline(args []q.Arg) (res float64) {
  deadline := -1.0
  for _, arg := range args {
    if arg.Name == "deadline" {
      deadline, _ = strconv.ParseFloat(arg.Value, 64)
    }
  }
  return deadline
}

// =============================================================================
// commitmentsToBytes - converts a slice of commitment structs to a byte array.
// =============================================================================
func commitmentsToBytes(commitments []Commitment) (res []byte, err error) {
  buf := new(bytes.Buffer)
  b, err := json.Marshal(commitments)
  if err != nil {
    return nil, err
  }
  err = binary.Write(buf, binary.BigEndian, &b)
  if err != nil {
    return nil, err
  }
  return buf.Bytes(), nil
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
