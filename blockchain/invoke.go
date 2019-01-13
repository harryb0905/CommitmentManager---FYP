package blockchain

import (
  "fmt"
  "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
  "time"
)

// Initialise a new commitment spec on the blockchain
func (setup *FabricSetup) InvokeInitCommitment(inpargs []string) (string, error) {

  // Prepare arguments
  var args []string
  args = append(args, "initCommitment")
  args = append(args, inpargs[0])
  args = append(args, inpargs[1])
  args = append(args, inpargs[2])
  args = append(args, inpargs[3])
  args = append(args, inpargs[4])

  eventID := "eventInvoke"

  // Add data that will be visible in the proposal, like a description of the invoke request
  transientDataMap := make(map[string][]byte)
  transientDataMap["result"] = []byte("Transient data in init commitment invoke")

  reg, notifier, err := setup.event.RegisterChaincodeEvent(setup.ChainCodeID, eventID)
  if err != nil {
    return "", err
  }
  defer setup.event.Unregister(reg)

  // Create a request (proposal) and send it
  response, err := setup.client.Execute(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3]), []byte(args[4]), []byte(args[5])}, TransientMap: transientDataMap})
  if err != nil {
    return "", fmt.Errorf("failed to move funds: %v", err)
  }

  // Wait for the result of the submission
  select {
  case ccEvent := <-notifier:
    fmt.Printf("Received CC event: %v\n", ccEvent)
  case <-time.After(time.Second * 20):
    return "", fmt.Errorf("did NOT receive CC event for eventId(%s) in init commitment", eventID)
  }

  return string(response.TransactionID), nil
}

// Add data to blockchain
func (setup *FabricSetup) InvokeInitCommitmentData(jsonStrs []string) (string, error) {

  // Prepare arguments
  argBytes := strArrToByteArr(jsonStrs)
  eventID := "eventInvoke"

  // Add data that will be visible in the proposal, like a description of the invoke request
  transientDataMap := make(map[string][]byte)
  transientDataMap["result"] = []byte("Transient data in init commitment data invoke")

  reg, notifier, err := setup.event.RegisterChaincodeEvent(setup.ChainCodeID, eventID)
  if err != nil {
    return "", err
  }
  defer setup.event.Unregister(reg)

  // Create a request (proposal) and send it
  response, err := setup.client.Execute(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: "initCommitmentData", Args: argBytes, TransientMap: transientDataMap})
  if err != nil {
    return "", fmt.Errorf("failed to move funds: %v", err)
  }

  // Wait for the result of the submission
  select {
  case ccEvent := <-notifier:
    fmt.Printf("Received CC event: %v\n", ccEvent)
  case <-time.After(time.Second * 20):
    return "", fmt.Errorf("did NOT receive CC event for eventId(%s) in init commitment data", eventID)
  }

  return string(response.TransactionID), nil
}

// Converts an array of strings to an array of byte arrays
func strArrToByteArr(strArr []string) (byteArr [][]byte) {
  output := make([][]byte, len(strArr))
  for i, v := range strArr {
    output[i] = []byte(v)
  }
  return output
}
