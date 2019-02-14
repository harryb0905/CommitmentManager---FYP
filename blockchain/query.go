package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// GetSpec - query the chaincode to get the state of a spec
func (setup *FabricSetup) GetSpec(name string) (string, error) {

  // Prepare arguments
  var args []string
  args = append(args, "getSpec")
  args = append(args, name)

  response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if err != nil {
    return "", fmt.Errorf("failed to query: %v", err)
  }

  return string(response.Payload), nil
}

// GetCreatedCommitments - query the chaincode to obtain created commitments
func (setup *FabricSetup) GetCreatedCommitments(comName string) (string, error) {

  // Prepare arguments
  var args []string
  args = append(args, "getCreatedCommitments")
  args = append(args, comName)

  response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if err != nil {
    return "", fmt.Errorf("failed to query: %v", err)
  }

  return string(response.Payload), nil
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
