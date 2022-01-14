package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Voter struct {
	ID      string            `json:"ID"`
	Ballots map[string]string `json:"Ballots"` // The ballots cast by this voter as a map of ElectionID -> BallotID
	Name    string            `json:"Name"`    // The name of this voter
}

func (s *SmartContract) GetVoter(ctx contractapi.TransactionContextInterface, id string) (*Voter, error) {
	JSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if JSON == nil {
		return nil, fmt.Errorf("the voter %s does not exist", id)
	}

	var voter Voter
	err = json.Unmarshal(JSON, &voter)
	if err != nil {
		return nil, err
	}

	return &voter, nil
}

func (s *SmartContract) NewVoter(ctx contractapi.TransactionContextInterface, id string, name string) error {

	voter := Voter{ID: id, Name: name}

	JSON, err := json.Marshal(voter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(voter.ID, JSON)
}

// Link a ballot to the voter's record
func (s *SmartContract) LogBallot(ctx contractapi.TransactionContextInterface, id string, ballot Ballot) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	voter, err := s.GetVoter(ctx, id)
	if err != nil {
		return err
	}

	if voter.Ballots == nil {
		voter.Ballots = make(map[string]string)
	}
	voter.Ballots[ballot.ElectionID] = ballot.ID

	JSON, err := json.Marshal(voter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, JSON)
}
