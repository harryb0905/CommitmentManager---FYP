package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// QueryCommitment query the chaincode to get the state of a commitment
func (setup *FabricSetup) QueryCommitment(name string) (string, error) {

  // Prepare arguments
  var args []string
  args = append(args, "getCommitment")
  args = append(args, name)

  response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
  if err != nil {
    return "", fmt.Errorf("failed to query: %v", err)
  }

  return string(response.Payload), nil
}

// RichQuery query the chaincode to perform an ad hoc rich query based on input
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
