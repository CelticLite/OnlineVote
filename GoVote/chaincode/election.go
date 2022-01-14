package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Election struct {
	ID               string   `json:"ID"`
	Name             string   `json:"Name"`             // The name of this election
	RegisteredVoters []string `json:"RegisteredVoters"` // The IDs of voters authorized to vote in this election
	VotableItems     []string `json:"VotableItems"`     // The IDs of the votable items in this election
}

// ReadElection returns the election stored in the world state with given id.
func (s *SmartContract) GetElection(ctx contractapi.TransactionContextInterface, id string) (*Election, error) {
	electionJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if electionJSON == nil {
		return nil, fmt.Errorf("the election %s does not exist", id)
	}

	var election Election
	err = json.Unmarshal(electionJSON, &election)
	if err != nil {
		return nil, err
	}

	return &election, nil
}

func (s *SmartContract) NewElection(ctx contractapi.TransactionContextInterface, id string, name string, voters []string, items []string) error {

	election := Election{ID: id, Name: name, RegisteredVoters: voters, VotableItems: items}

	JSON, err := json.Marshal(election)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, JSON)
}
