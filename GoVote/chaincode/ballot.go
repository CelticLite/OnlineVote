package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Ballot struct {
	ID           string   `json:"ID"`
	VoterID      string   `json:"VoterID"`      // The ID of the voter who cast this ballot
	ElectionID   string   `json:"ElectionID"`   // The ID of the election this ballot is for
	VotableItems []string `json:"VotableItems"` // The voter's choices (as VotableItem IDs) in order of preference
}

func (s *SmartContract) GetBallot(ctx contractapi.TransactionContextInterface, id string) (*Ballot, error) {
	JSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if JSON == nil {
		return nil, fmt.Errorf("the ballot %s does not exist", id)
	}

	var ballot Ballot
	err = json.Unmarshal(JSON, &ballot)
	if err != nil {
		return nil, err
	}

	return &ballot, nil
}

// Checks the validity of a Ballot
// Returns an error if the check failed or if the ballot is invalid
func (s *SmartContract) ValidateBallot(ctx contractapi.TransactionContextInterface, ballot Ballot) error {
	// Check that the voter exists and has not already cast a ballot
	voter, err := s.GetVoter(ctx, ballot.VoterID)
	if err != nil {
		return err
	}
	if _, voted := voter.Ballots[ballot.ElectionID]; voted {
		return fmt.Errorf("the voter has already cast a ballot for this election")
	}

	// Check that the voter is registered for the election
	election, err := s.GetElection(ctx, ballot.ElectionID)
	if err != nil {
		return err
	}
	registered := false
	for _, v := range election.RegisteredVoters {
		if ballot.VoterID == v {
			registered = true
			break
		}
	}
	if !registered {
		return fmt.Errorf("voter %s is not registered for the election %s", ballot.VoterID, ballot.ElectionID)
	}

	// Check that only valid VotableItems have been selected and that each exists exactly once
	if len(election.VotableItems) != len(ballot.VotableItems) {
		return fmt.Errorf("invalid number of selected VotableItems")
	}

	for _, eitem := range election.VotableItems {
		onBallot := false
		for _, bitem := range ballot.VotableItems {
			if eitem == bitem {
				onBallot = true
				break
			}
		}
		if !onBallot {
			return fmt.Errorf("invalid selection of VotableItems")
		}
	}

	return nil
}

// Creates a new Ballot and commits it to the ledger if it is valid
func (s *SmartContract) NewBallot(ctx contractapi.TransactionContextInterface, voterID string, electionID string, votableItems []string) (*Ballot, error) {

	ballot := Ballot{VoterID: voterID, ElectionID: electionID, VotableItems: votableItems}
	ballot.ID = fmt.Sprintf("%s-%s", voterID, electionID)

	err := s.ValidateBallot(ctx, ballot)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ballot: %s", err)
	}

	JSON, err := json.Marshal(ballot)
	if err != nil {
		return nil, err
	}

	return &ballot, ctx.GetStub().PutState(ballot.ID, JSON)
}
