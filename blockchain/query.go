package blockchain

import (
	"fmt"
  "encoding/json"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

type Commitment struct {
  ComID    string
  States []ComState
}

type ComState struct {
  Name  string
  Data  map[string]interface{}
}

type CommitmentMeta struct {
  Name     string  `json:"name"`
  Source   string  `json:"source"`
  Summary  string  `json:"summary"`
}

// GetSpec - query the chaincode to get the state of a spec
func (setup *FabricSetup) GetSpec(name string) (res *CommitmentMeta, err error) {

  // Prepare results
  com := &CommitmentMeta{}

  // Prepare arguments
  var args []string
  args = append(args, "getSpec")
  args = append(args, name)

  response, er := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if er != nil {
    return com, fmt.Errorf("failed to query: %v", er)
  }

  json.Unmarshal([]byte(response.Payload), &com)
  return com, nil
}

// GetCreatedCommitments - query the chaincode to obtain created commitments
func (setup *FabricSetup) GetCreatedCommitments(comName string) (coms[] Commitment, err error) {

  // Prepare results
  commitments := []Commitment{}

  // Prepare arguments
  var args []string
  args = append(args, "getCreatedCommitments")
  args = append(args, comName)

  response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if err != nil {
    return commitments, fmt.Errorf("failed to query: %v", err)
  }

  json.Unmarshal([]byte(response.Payload), &commitments)
  return commitments, nil
}

// RichQuery - query the chaincode to perform an ad hoc rich query based on input
func (setup *FabricSetup) RichQuery(query string) (string, error) {

  // Prepare arguments
  var args []string
  args = append(args, "richQuery")
  args = append(args, query)

  response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if err != nil {
    return "", fmt.Errorf("failed to perform rich query: %v", err)
  }

  return string(response.Payload), nil
}
