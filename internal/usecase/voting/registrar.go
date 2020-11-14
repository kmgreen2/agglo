package voting

import (
	"fmt"
	gUuid "github.com/google/uuid"
	"strings"
)

type VoterRecord struct {
	voterID gUuid.UUID
	electionUuidSignature []byte
}

type RegistrarBackChannel interface {
	// Closes loop that user has registered for an electionUuid
	ElectionUUIDCommit(voterElectionUuid gUuid.UUID, voterElectionUuidSignature []byte) error
}

type Registrar struct {
	voters map[gUuid.UUID]*VoterRecord	 	// All people registered to vote in this election
	people map[string]*Person				// All people registered to vote
	registeredPeople map[string]gUuid.UUID  // All people that showed intention to vote in this election
}

func NewRegistrar(registeredPeople []*Person) *Registrar {
	people := make(map[string]*Person)
	for _, person := range registeredPeople {
		people[person.id] = person
	}
	return &Registrar{
		voters: make(map[gUuid.UUID]*VoterRecord),
		people: people,
	}
}

func (registrar *Registrar) getVoterID(personID, secret string) (gUuid.UUID, error) {
	if person, ok := registrar.people[personID]; ok {
		if strings.Compare(secret, person.secret) == 0 {
			if voterID, ok := registrar.registeredPeople[personID]; ok {
				return voterID, nil
			}
			voterID := gUuid.New()
			registrar.registeredPeople[personID] = voterID
			return voterID, nil
		}
	}
	return gUuid.Nil, fmt.Errorf("Cannot find person with ID '%s'", personID)
}

// ElectionRegister will allocate a unique election ID for this voter
func (registrar *Registrar) ElectionRegister(personID, secret string) (gUuid.UUID, error) {
	return registrar.getVoterID(personID, secret)
}


// PrepareToVote will return the signature of the electionUuid, which is needed to vote
func (registrar *Registrar) PrepareToVote(personID, secret string) ([]byte, error) {
	voterID, err := registrar.getVoterID(personID, secret)
	if err != nil {
		return nil, err
	}
	if voterRecord, ok := registrar.voters[voterID]; ok {
		return voterRecord.electionUuidSignature, nil
	}
	return nil, fmt.Errorf("Cannot find election ID generator receipt for '%s'", personID)
}
