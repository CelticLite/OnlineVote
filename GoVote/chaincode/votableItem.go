package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type VotableItem struct {
	ID          string `json:"ID"`          // The ID of this votable item
	Name        string `json:"Name"`        // The name of this votable item
	Description string `json:"Description"` // The description of this votable item
}

func (s *SmartContract) NewVotableItem(ctx contractapi.TransactionContextInterface, id string, name string, desc string) error {

	item := VotableItem{ID: id, Name: name, Description: desc}

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, itemJSON)
}

func (s *SmartContract) GetVotableItem(ctx contractapi.TransactionContextInterface, id string) (*VotableItem, error) {
	itemJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if itemJSON == nil {
		return nil, fmt.Errorf("the votableitem %s does not exist", id)
	}

	var item VotableItem
	err = json.Unmarshal(itemJSON, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}
