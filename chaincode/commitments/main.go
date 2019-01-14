package quark

import (
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	"github.com/chainHero/heroes-service/blockchain"
	q "github.com/chainHero/heroes-service/quark"
	p "github.com/chainHero/heroes-service/quark/parser"
)

const (
  GetEventQuery = "{\"selector\":{\"docType\":\"%s\"}}"
	TimeFormat = "Mon Jan _2 15:04:05 2006"
)

type State struct {
	Name 			string
	Responses []QueryResponse
}

type QueryResponse struct {
  Key     string
  Record  map[string]interface{}
}

type Commitment struct {
  Name     string  `json:"name"`
  Source   string  `json:"source"`
  Summary  string  `json:"summary"`
}

// =========================== GET CREATED COMMITMENTS ===================================
//
// Obtains all created commitments based on a given commitment/spec name.
// A commitment is created if it exists on the blockchain CouchDB database.
//
// =======================================================================================
func GetCreatedCommitments(comName string, fab *blockchain.FabricSetup) (states []State, com Commitment, err error) {
	states = []State {
    State { Name: "Created", Responses: nil, },
		State { Name: "Detached", Responses: nil, },
		State { Name: "Discharged", Responses: nil, },
	}

	// Get all commitment details
  com, spec := getCommitmentDetails(comName, fab)
	events := []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name}

	if (len(events) > 0) {
		createEvent := events[0]
		query := fmt.Sprintf(GetEventQuery, createEvent)
		// Perform query to get created commitment results
	  response, err := fab.RichQuery(query)
	  if err != nil {
	    return nil, com, err;
	  } else {
	    // Unmarshal JSON
	    responses := []QueryResponse{}
	    err = json.Unmarshal([]byte(response), &responses)
			states[0].Responses = responses
			fmt.Println(responses)
	    return states, com, err;
	  }
	}
	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

// =========================== GET DETACHED COMMITMENTS ==================================
//
// Obtains all detached commitments based on a given commitment/spec name.
// A commitment is detached if the created event exists on the blockchain CouchDB database
// and the detached event has occured within the specified deadline.
// If the commitment isn't detached and the deadline has exceeded, the commitment expires.
//
// =======================================================================================
func GetDetachedCommitments(comName string, fab *blockchain.FabricSetup) (states []State, com Commitment, err error) {
	states = []State {
    State { Name: "Created", Responses: nil, },
		State { Name: "Detached", Responses: nil, },
		State { Name: "Discharged", Responses: nil, },
	}
	// 1. Check if this commitment has been created (can't be detached if not already created)
	comStates, _, _ := GetCreatedCommitments(comName, fab)
	createdComs := comStates[0].Responses

	// 2. Get created event name

	// 2. Get keys from created results slice
	for _, element := range createdComs {
		fmt.Println("key:", element.Key)
	}

	// 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	com, spec := getCommitmentDetails(comName, fab)
	events := []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name}
	detachArgs := spec.DetachEvent.Args
	spew.Dump(spec)

	fmt.Println("detach args:", detachArgs)

	// Get deadline value
	deadline := 0
	for _, arg := range detachArgs {
    if arg.Name == "deadline" {
			deadline, _ = strconv.Atoi(arg.Value)
    }
	}

	if (len(events) > 0) {
		detachEvent := events[1]
		query := fmt.Sprintf(GetEventQuery, detachEvent)

		fmt.Println("detach event:", detachEvent)

		response, err := fab.RichQuery(query)
		fmt.Println("response:", response)
		if err != nil {
			// return nil, com, err;
		} else {
			// Unmarshal JSON
			// 6. Obtain results into struct and return a slice of query responses to output in HTML table
			responses := []QueryResponse{}
			_ = json.Unmarshal([]byte(response), &responses)

			// For each commitment, perform date check with deadline and event date from commitment
			// 4. Perform Go time arithmetic on deadline (e.g. deadline=5 means payment must occur within 5 days of the offer being created)
			for i, com := range responses {
				// 5. If Pay record exists and that timestamp is within a period of 5 days or less from offer being created, this commitment is detached
				// Get created commitment that corresponds to this detached commitment
				for _, createdCom := range createdComs {
					if (createdCom.Record["comID"].(string) == com.Record["comID"].(string)) {
						createdDateStr := com.Record["date"].(string)
						detachedDateStr := createdCom.Record["date"].(string)
						// If detached event date isn't within the specified deadline, remove from results
						if (!isDateWithinDeadline(createdDateStr, detachedDateStr, deadline)) {
							responses[i] = responses[len(responses)-1]
						  responses = responses[:len(responses)-1]
						}
					}
				}
			}

			states[1].Responses = responses

			return states, com, err;
		}
	}

	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

// =========================== GET EXPIRED COMMITMENTS ===================================
//
// Obtains all expired commitments based on a given commitment/spec name
//
// =======================================================================================
func GetExpiredCommitments(comName string, fab *blockchain.FabricSetup) (results []QueryResponse, com Commitment, err error) {
	// 1. Check if this commitment has been created (can't be detached if not already created)
	// createdComs, _, _ := GetCreatedCommitments(comName, fab)
	//
	// // 2. Get created event name
	//
	//
	// // 2. Get keys from created results slice
	// for _, element := range createdComs {
	// 	fmt.Println("key:", element.Key)
	// }
	//
	// // 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	// com, spec := getCommitmentDetails(comName, fab)
	// events := []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name}
	//
	// if (len(events) > 0) {
	// 	detachEvent := events[1]
	// 	query := fmt.Sprintf(GetEventQuery, detachEvent)
	//
	// 	fmt.Println("detach event:", detachEvent)
	//
	// 	response, err := fab.RichQuery(query)
	// 	fmt.Println("response:", response)
	// 	if err != nil {
	// 		// return nil, com, err;
	// 	} else {
	// 		// Unmarshal JSON
	// 		results := []QueryResponse{}
	// 		_ = json.Unmarshal([]byte(response), &results)
	// 		fmt.Println("detach results:", results)
	// 		return results, com, err;
	// 	}
	// }

	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

// =========================== GET DISCHARGED COMMITMENTS ==================================
//
// Obtains all discharged commitments based on a given commitment/spec name
//
// =========================================================================================
func GetDischargedCommitments(comName string, fab *blockchain.FabricSetup) (results []QueryResponse, com Commitment, err error) {
	// 1. Check if this commitment has been created (can't be detached if not already created)
	// createdComs, _, _ := GetCreatedCommitments(comName, fab)
	//
	// // 2. Get created event name
	//
	//
	// // 2. Get keys from created results slice
	// for _, element := range createdComs {
	// 	fmt.Println("key:", element.Key)
	// }
	//
	// // 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	// com, spec := getCommitmentDetails(comName, fab)
	// events := []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name}
	//
	// if (len(events) > 0) {
	// 	detachEvent := events[1]
	// 	query := fmt.Sprintf(GetEventQuery, detachEvent)
	//
	// 	fmt.Println("detach event:", detachEvent)
	//
	// 	response, err := fab.RichQuery(query)
	// 	fmt.Println("response:", response)
	// 	if err != nil {
	// 		// return nil, com, err;
	// 	} else {
	// 		// Unmarshal JSON
	// 		results := []QueryResponse{}
	// 		_ = json.Unmarshal([]byte(response), &results)
	// 		fmt.Println("detach results:", results)
	// 		return results, com, err;
	// 	}
	// }

	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

//
// =========================== GET VIOLATED COMMITMENTS ====================================
//
// Obtains all violated commitments based on a given commitment/spec name
//
// =========================================================================================
func GetViolatedCommitments(comName string, fab *blockchain.FabricSetup) (results []QueryResponse, com Commitment, err error) {
	// 1. Check if this commitment has been created (can't be detached if not already created)
	// createdComs, _, _ := GetCreatedCommitments(comName, fab)
	//
	// // 2. Get created event name
	//
	//
	// // 2. Get keys from created results slice
	// for _, element := range createdComs {
	// 	fmt.Println("key:", element.Key)
	// }
	//
	// // 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	// com, spec := getCommitmentDetails(comName, fab)
	// events := []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name}
	//
	// if (len(events) > 0) {
	// 	detachEvent := events[1]
	// 	query := fmt.Sprintf(GetEventQuery, detachEvent)
	//
	// 	fmt.Println("detach event:", detachEvent)
	//
	// 	response, err := fab.RichQuery(query)
	// 	fmt.Println("response:", response)
	// 	if err != nil {
	// 		// return nil, com, err;
	// 	} else {
	// 		// Unmarshal JSON
	// 		results := []QueryResponse{}
	// 		_ = json.Unmarshal([]byte(response), &results)
	// 		fmt.Println("detach results:", results)
	// 		return results, com, err;
	// 	}
	// }

	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

func isDateWithinDeadline(createdDateStr string, detachedDateStr string, deadline int) (within bool) {
	createEventDate, _ := time.Parse(TimeFormat, createdDateStr)
	detachEventDate, _ := time.Parse(TimeFormat, detachedDateStr)
	daysDiff := int(createEventDate.Sub(detachEventDate).Hours() / 24)

	if (daysDiff >= deadline) {
		return false
	} else {
		return true
	}
}

// Obtains the events for a given commitment (e.g. Offer, Pay, Delivery)
func getCommitmentDetails(comName string, fab *blockchain.FabricSetup) (res Commitment, spec *p.Spec) {

	// Obtain commitment from CouchDB based on the comName
	response, _ := fab.QueryCommitment(comName)

	// Unmarshal JSON into structure and obtain source code
	com := Commitment{}
	json.Unmarshal([]byte(response), &com)

	// Compile specification (using custom built compiler) to obtain events
	spec, _ = q.Parse(com.Source)
	fmt.Println("RESPONSE:", )

	// Extract event names and append to slice
	if (spec != nil) {
	  return com, spec;
	}
	return com, nil
}
