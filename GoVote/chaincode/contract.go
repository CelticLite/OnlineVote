package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	voters := []Voter{
		{ID: "voter1", Name: "Alan Turing"},
		{ID: "voter2", Name: "Ada Lovelace"},
		{ID: "voter3", Name: "Unregistered"},
	}

	votableItems := []VotableItem{
		{ID: "item1", Name: "A. Lincon", Description: "Candidate for president."},
		{ID: "item2", Name: "R. Nixon", Description: "Candidate for president."},
		{ID: "item3", Name: "N. Real", Description: "Invalid Option."},
	}

	election := Election{ID: "election1", Name: "Sample Presidential Election", RegisteredVoters: []string{"voter1", "voter2"}, VotableItems: []string{"item1", "item2"}}

	for _, voter := range voters {
		voterJSON, err := json.Marshal(voter)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(voter.ID, voterJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, item := range votableItems {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(item.ID, itemJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	electionJSON, err := json.Marshal(election)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(election.ID, electionJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// Returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// Calculate the results of an election and return a map of vitem -> count as a string
func (s *SmartContract) GetResults(ctx contractapi.TransactionContextInterface, electionID string) (string, error) {
	results := map[string]int{}
	election, err := s.GetElection(ctx, electionID)
	if err != nil {
		return "", err
	}

	// Initialize counts to zero
	for _, itemID := range election.VotableItems {
		results[itemID] = 0
	}

	// Count ballots
	for _, voterID := range election.RegisteredVoters {
		voter, err := s.GetVoter(ctx, voterID)
		if err != nil {
			return "", err
		}

		ballotID, exists := voter.Ballots[electionID]
		if exists {
			ballot, err := s.GetBallot(ctx, ballotID)
			if err != nil {
				return "", err
			}
			for i, itemID := range ballot.VotableItems {
				results[itemID] += len(election.VotableItems) - i
			}
		}
	}

	return fmt.Sprint(results), nil
}

// Create ballot and link to voter
func (s *SmartContract) CastVote(ctx contractapi.TransactionContextInterface, voterID string, electionID string, votableItems []string) error {
	// Create the ballot
	ballot, err := s.NewBallot(ctx, voterID, electionID, votableItems)
	if err != nil {
		return fmt.Errorf("failed to create ballot: %s", err)
	}

	// Link the ballot to the voter
	err = s.LogBallot(ctx, voterID, *ballot)
	if err != nil {
		return fmt.Errorf("failed to link ballot to voter: %s", err)
	}

	return nil
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]string, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		assets = append(assets, string(queryResponse.Value))
	}

	return assets, nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
