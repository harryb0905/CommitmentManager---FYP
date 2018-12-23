package quark

import (
	"fmt"
	"encoding/json"
	"github.com/chainHero/heroes-service/blockchain"
	u "github.com/chainHero/heroes-service/utils/queries"
	q "github.com/chainHero/heroes-service/quark"
)

type QueryResponse struct {
  Key     string
  Record  map[string]interface{}
}

type CommitmentResponse struct {
  Name     string  `json:"name"`
  Source   string  `json:"source"`
  Summary  string  `json:"summary"`
}

// Obtains all created commitments based on a given commitment/spec name
func GetCreatedCommitments(comName string, fab *blockchain.FabricSetup) (results []QueryResponse, com CommitmentResponse, err error)  {
  com, events := getCommitmentDetails(comName, fab)
	if (len(events) > 0) {
		createEvent := events[0]
		query := fmt.Sprintf(u.CreatedQuery, createEvent)
	  response, err := fab.RichQuery(query)
	  if err != nil {
	    return nil, com, err;
	  } else {
	    // Unmarshal JSON
	    results := []QueryResponse{}
	    err = json.Unmarshal([]byte(response), &results)
	    return results, com, err;
	  }
	}
	return nil, com, fmt.Errorf("Couldn't get events for this commitment")
}

// Obtains all detached commitments based on a given commitment/spec name
// Commitment is detached if payment has been made in specified deadline
// If this commitment is not detached and the deadline has been exceeded, the commitment expires
func GetDetachedCommitments(comName string, fab *blockchain.FabricSetup) {
	// 1. Check if this commitment has been created (can't be detached if not already created)

	// 2. If created, extract the deadline value from the spec source (e.g. deadline=5)
	_, events := getCommitmentDetails(comName, fab)
	detachEvent := events[1]
	query := fmt.Sprintf(u.CreatedQuery, detachEvent)

	response, err := fab.RichQuery(query)
	if err != nil {
		// return nil, com, err;
	} else {
		// Unmarshal JSON
		results := []QueryResponse{}
		_ = json.Unmarshal([]byte(response), &results)
		fmt.Println("detach results:", results)
	}

	// 3. Perform Go time arithmetic on deadline (e.g. deadline=5 means payment must occur within 5 days of the offer being created)

	// 4. If Pay record exists and that timestamp is within a period of 5 days or less from offer being created, this commitment is detached


	// 5. Obtain results into struct and return a slice of query responses to output in HTML table


}

// Obtains all expired commitments based on a given commitment/spec name
func GetExpiredCommitments(comName string, fab *blockchain.FabricSetup)  {

}

// Obtains the events for a given commitment (e.g. Offer, Pay, Delivery etc)
func getCommitmentDetails(comName string, fab *blockchain.FabricSetup) (res CommitmentResponse, events[]string) {

	// Obtain commitment from CouchDB based on the comName
	response, _ := fab.QueryCommitment(comName)

	// Unmarshal JSON into structure and obtain source code
	com := CommitmentResponse{}
	json.Unmarshal([]byte(response), &com)

	// Parse specification to obtain events
	spec, _ := q.Parse(com.Source)

	// Extract event names and append to slice
	if (spec != nil) {
		events = []string{spec.CreateEvent.Name, spec.DetachEvent.Name, spec.DischargeEvent.Name} // may need .Args here for more querying impl...
	  return com, events;
	}
	return com, nil
}
