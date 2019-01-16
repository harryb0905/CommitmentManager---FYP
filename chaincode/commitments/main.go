package quark

import (
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	// "github.com/davecgh/go-spew/spew"
	"github.com/chainHero/heroes-service/blockchain"
	q "github.com/chainHero/heroes-service/quark"
	p "github.com/chainHero/heroes-service/quark/parser"
)

const (
  GetEventQuery = "{\"selector\":{\"docType\":\"%s\"}}"
	TimeFormat = "Mon Jan _2 15:04:05 2006"
)

type Commitment struct {
	ComID    string
	States []ComState
}

type ComState struct {
	Name 	string
	Data  map[string]interface{}
}

type QueryResponse struct {
  Key     string
  Record  map[string]interface{}
}

type CommitmentMeta struct {
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
func GetCreatedCommitments(comName string, fab *blockchain.FabricSetup) (commitments []Commitment, com CommitmentMeta, err error) {

	// Get all commitment details (inc. event names)
  com, spec := getCommitmentDetails(comName, fab)
	createEvent := spec.CreateEvent.Name

	// Format query and perform query to get created commitment results
	query := fmt.Sprintf(GetEventQuery, createEvent)
  response, err := fab.RichQuery(query)

  if err != nil {
    return nil, com, err;
  } else {
    // Unmarshal JSON response from query
		responses := []QueryResponse{}
		err = json.Unmarshal([]byte(response), &responses)

		// Create commitments from responses with data per commitment state
		commitments = []Commitment{}
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
    return commitments, com, err;
  }

	return nil, com, fmt.Errorf("Couldn't get %s created commitments", comName)
}

// func getRecordsByComID(comID string, ) (records []map[string]interface{}) {
//
// }

// =========================== GET DETACHED COMMITMENTS ==================================
//
// Obtains all detached commitments based on a given commitment/spec name.
// A commitment is detached if the created event exists on the blockchain CouchDB database
// and the detached event has occured within the specified deadline.
// If the commitment isn't detached and the deadline has exceeded, the commitment expires.
//
// =======================================================================================
func GetDetachedCommitments(comName string, wantExpired bool, fab *blockchain.FabricSetup) (commitments []Commitment, com CommitmentMeta, err error) {

	// 1. Check if this commitment has been created (can't be detached if not already created)
	createdComs, com, err := GetCreatedCommitments(comName, fab)

	// 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	com, spec := getCommitmentDetails(comName, fab)
	detachEvent := spec.DetachEvent.Name
	detachArgs := spec.DetachEvent.Args
	// spew.Dump(spec)

	// Get deadline value
	deadline := 0
	for _, arg := range detachArgs {
    if arg.Name == "deadline" {
			deadline, _ = strconv.Atoi(arg.Value)
    }
	}

	// Format and perform query
	query := fmt.Sprintf(GetEventQuery, detachEvent)
	response, err := fab.RichQuery(query)

	if err != nil {
		// return nil, com, err;
	} else {
		// 6. Unmarshal JSON response from query
		responses := []QueryResponse{}
		_ = json.Unmarshal([]byte(response), &responses)

		// Create commitments from responses with data per commitment state
		commitments = []Commitment{}

		// 4. Date checks with deadlines on detached results
		for _, comRes := range responses {

			// 5. If Pay record exists and that timestamp is within a period of 5 days or less from offer being created, this commitment is detached.
			for _, createdCom := range createdComs {
				// Get created commitment that corresponds to this detached commitment
				if (createdCom.ComID == comRes.Record["comID"].(string)) {
					// Extract date for checking deadline
					createdDateStr := createdCom.States[0].Data["date"].(string)
					detachedDateStr := comRes.Record["date"].(string)

					// If detached event date is within specified deadline, include in results
					withinDeadline := isDateWithinDeadline(createdDateStr, detachedDateStr, deadline)
					if ((withinDeadline && !wantExpired) || (!withinDeadline && wantExpired)) {
						commitments = append(commitments,
							Commitment{
								ComID: comRes.Record["comID"].(string),
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
				}
			}
		}
		return commitments, com, err;
	}

	return nil, com, fmt.Errorf("Couldn't get %s detached commitments", comName)
}

// =========================== GET EXPIRED COMMITMENTS ===================================
//
// Obtains all expired commitments based on a given commitment/spec name
//
// =======================================================================================
func GetExpiredCommitments(comName string, fab *blockchain.FabricSetup) (commitments []Commitment, com CommitmentMeta, err error) {
	expiredComs, com, err := GetDetachedCommitments(comName, true, fab)
	if (err == nil) {
		return expiredComs, com, err;
	}
	return nil, com, fmt.Errorf("Couldn't get %s expired commitments", comName)
}

// =========================== GET DISCHARGED COMMITMENTS ==================================
//
// Obtains all discharged commitments based on a given commitment/spec name
//
// =========================================================================================
func GetDischargedCommitments(comName string, wantExpired bool, fab *blockchain.FabricSetup) (commitments []Commitment, com CommitmentMeta, err error) {
	// 1. Check if this commitment has been created (can't be detached if not already created)
	detachedComs, com, err := GetDetachedCommitments(comName, false, fab)

	// 3. If created, extract the deadline value from the spec source (e.g. deadline=5)
	com, spec := getCommitmentDetails(comName, fab)
	dischargeEvent := spec.DischargeEvent.Name
	dischargeArgs := spec.DischargeEvent.Args
	// spew.Dump(spec)

	// Get deadline value
	deadline := 0
	for _, arg := range dischargeArgs {
    if arg.Name == "deadline" {
			deadline, _ = strconv.Atoi(arg.Value)
    }
	}

	// Format and perform query
	query := fmt.Sprintf(GetEventQuery, dischargeEvent)
	response, err := fab.RichQuery(query)

	if err != nil {
		// return nil, com, err;
	} else {
		// 6. Unmarshal JSON response from query
		responses := []QueryResponse{}
		_ = json.Unmarshal([]byte(response), &responses)

		// Create commitments from responses with data per commitment state
		commitments = []Commitment{}


		// 4. Date checks with deadlines on detached results
		for _, comRes := range responses {
			fmt.Println("comRes:", comRes)

			// 5. If Pay record exists and that timestamp is within a period of 5 days or less from offer being created, this commitment is detached.
			for _, detachedCom := range detachedComs {
				fmt.Println(detachedCom.ComID, comRes.Record["comID"].(string))

				// Get created commitment that corresponds to this detached commitment
				if (detachedCom.ComID == comRes.Record["comID"].(string)) {

					// Extract date for checking deadline
					createdDateStr := detachedCom.States[0].Data["date"].(string)
					dischargedDateStr := comRes.Record["date"].(string)

					// If detached event date is within specified deadline, include in results
					withinDeadline := isDateWithinDeadline(createdDateStr, dischargedDateStr, deadline)
					if ((withinDeadline && !wantExpired) || (!withinDeadline && wantExpired)) {
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
										Data: comRes.Record,
									},
									ComState{
										Name: "Discharged",
										Data: comRes.Record,
									},
								},
							},
						)
					}
				}
			}
		}
		return commitments, com, err;
	}

	return nil, com, fmt.Errorf("Couldn't get %s detached commitments", comName)
}

// =========================== GET VIOLATED COMMITMENTS ====================================
//
// Obtains all violated commitments based on a given commitment/spec name
//
// =========================================================================================
func GetViolatedCommitments(comName string, fab *blockchain.FabricSetup) (commitments []Commitment, com CommitmentMeta, err error) {
	violatedComs, com, err := GetDischargedCommitments(comName, true, fab)
	if (err == nil) {
		return violatedComs, com, err;
	}
	return nil, com, fmt.Errorf("Couldn't get %s violated commitments", comName)
}

// Perform Go time arithmetic on deadline (e.g. deadline=5 means payment must occur within 5 days of the offer being created)
func isDateWithinDeadline(createdDateStr string, detachedDateStr string, deadline int) (within bool) {
	createEventDate, _ := time.Parse(TimeFormat, createdDateStr)
	detachEventDate, _ := time.Parse(TimeFormat, detachedDateStr)
	daysDiff := int(detachEventDate.Sub(createEventDate).Hours() / 24)

	if (daysDiff >= deadline) {
		return false
	} else {
		return true
	}
}

// Obtains the events for a given commitment (e.g. Offer, Pay, Delivery)
func getCommitmentDetails(comName string, fab *blockchain.FabricSetup) (res CommitmentMeta, spec *p.Spec) {

	// Obtain commitment from CouchDB based on the comName
	response, _ := fab.QueryCommitment(comName)

	// Unmarshal JSON into structure and obtain source code
	com := CommitmentMeta{}
	json.Unmarshal([]byte(response), &com)

	// Compile specification (using custom built compiler) to obtain events
	spec, _ = q.Parse(com.Source)

	// Extract event names and append to slice
	if (spec != nil) {
	  return com, spec;
	}
	return com, nil
}
