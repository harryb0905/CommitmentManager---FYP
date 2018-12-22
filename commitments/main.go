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
		query := fmt.Sprintf(u.CreatedQuery, events[0])
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
