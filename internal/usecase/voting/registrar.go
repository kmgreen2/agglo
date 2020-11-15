package voting

import (
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"strings"
)

type VoterRecord struct {
	voterID gUuid.UUID
	electionUuidSignature []byte
}

type Poll struct {
	publicInfo *PublicInfo
	ballot *Ballot
}

type RegistrarBackChannel interface {
	// Closes loop that user has registered for an electionUuid
	ElectionUUIDCommit(voterId gUuid.UUID, voterElectionUuidSignature []byte) error
}

type Registrar struct {
	voters map[gUuid.UUID]*VoterRecord	 		// All people registered to vote in this election
	people map[string]*Person					// All people registered to vote
	registeredPeople map[string]gUuid.UUID  	// All people that showed intention to vote in this election
	registeredVoterIDs map[gUuid.UUID]string    // All people that showed intention to vote in this election
	polls []*Poll								// Voter can optionally supply a poll when they register intent to vote
}

func NewRegistrar(registeredPeople []*Person) *Registrar {
	people := make(map[string]*Person)
	for _, person := range registeredPeople {
		people[person.ID] = person
	}
	return &Registrar{
		voters: make(map[gUuid.UUID]*VoterRecord),
		people: people,
		registeredPeople: make(map[string]gUuid.UUID),
		registeredVoterIDs: make(map[gUuid.UUID]string),
		polls: make([]*Poll, 0),
	}
}

// ListPeople will return a list of all IDs for those eligable to vote
func (registrar* Registrar) ListPeople() ([]string, error) {
	allPeople := make([]string, 0)
	for _, person := range registrar.people {
		allPeople = append(allPeople, person.ID)
	}
	return allPeople, nil
}

// ListRegistered will return a list of all IDs for those registered to vote in the election
func (registrar* Registrar) ListRegistered() ([]string, error) {
	allPeople := make([]string, 0)
	for personID, _ := range registrar.registeredPeople {
		allPeople = append(allPeople, personID)
	}
	return allPeople, nil
}

func (registrar *Registrar) getVoterID(personID, secret string) (gUuid.UUID, error) {
	if person, ok := registrar.people[personID]; ok {
		if strings.Compare(secret, person.Secret) == 0 {
			if voterID, ok := registrar.registeredPeople[personID]; ok {
				return voterID, nil
			}
			return gUuid.Nil, NewVoterIDNotFoundError(fmt.Sprintf("Cannot find voterID for person with ID '%s'",
				personID))
		}
	}
	return gUuid.Nil, NewUnauthorizedError(fmt.Sprintf("Cannot find person or secret does not match for ID '%s'",
		personID))
}

// ElectionRegister will allocate a unique election ID for this voter
func (registrar *Registrar) ElectionRegister(personID, secret string) (gUuid.UUID, error) {
	voterID, err := registrar.getVoterID(personID, secret)

	if errors.Is(err, &VoterIDNotFoundError{}) {
		voterID = gUuid.New()
		registrar.registeredPeople[personID] = voterID
		registrar.registeredVoterIDs[voterID] = personID
	} else if err != nil {
		return gUuid.Nil, err
	}
	return voterID, nil
}

// PrepareToVote will return the signature of the electionUuid, which is needed to vote
func (registrar *Registrar) PrepareToVote(personID, secret string, pollBallot *Ballot) ([]byte, error) {
	voterID, err := registrar.getVoterID(personID, secret)
	if err != nil {
		return nil, err
	}
	if pollBallot != nil {
		person := registrar.people[personID]
		registrar.polls = append(registrar.polls, &Poll{
			publicInfo: person.PublicInfo,
			ballot: pollBallot,
		})
	}
	if voterRecord, ok := registrar.voters[voterID]; ok {
		return voterRecord.electionUuidSignature, nil
	}
	return nil, fmt.Errorf("Cannot find election ID generator receipt for '%s'", personID)
}

// ElectionUUIDCommit will associate a electionUuid signature with a voterID
// ToDo(KMG): VoterID should be blinded, so the Generator does not have the mapping of voterID to electionUuid
// The only entity that should have voterID and electionUuid is the voter
func (registrar *Registrar) ElectionUUIDCommit(voterID gUuid.UUID, voterElectionUuidSignature []byte) error {
	if _, ok := registrar.voters[voterID]; ok {
		return fmt.Errorf("Election ID already exists for voter with ID: %s", voterID.String())
	}
	registrar.voters[voterID] = &VoterRecord{
		voterID: voterID,
		electionUuidSignature: voterElectionUuidSignature,
	}

	return nil
}