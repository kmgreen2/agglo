package voting

import (
	"bytes"
	"encoding/gob"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/storage"
	"io/ioutil"
	"strings"
	"time"
)

type LedgerBackChannel interface {
	GeneratorAuthenticator(authenticator crypto.Authenticator) error
	VoterAuthenticator(voterElectionUuid gUuid.UUID, authenticator crypto.Authenticator) error
}

type RaceID string

type Race struct {
	Choices map[string]bool
	MaxChoices int
}

func NewRace(choices []string, maxChoices int) *Race {
	choiceMap := make(map[string]bool)
	for _, choice := range choices {
		choiceMap[choice] = false
	}
	return &Race{
		Choices: choiceMap,
		MaxChoices: maxChoices,
	}
}

type Ballot struct {
	Races map[RaceID]*Race
}

type RaceResult struct {
	results map[string]int
}

type Tally struct {
	raceResults map[RaceID]*RaceResult
}

func SerializeBallot(ballot *Ballot) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(ballot)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func DeerializeBallot(ballotBytes []byte) (*Ballot, error) {
	byteBuffer := bytes.NewBuffer(ballotBytes)
	gEncoder := gob.NewDecoder(byteBuffer)
	ballot := &Ballot{}
	err := gEncoder.Decode(ballot)
	if err != nil {
		return nil, err
	}
	return ballot, nil
}

func (ballot *Ballot) InvalidRaces() []RaceID {
	invalidRaces := make([]RaceID, 0)
	for raceID, race := range ballot.Races {
		i := 0
		for _, chosen := range race.Choices {
			if chosen {
				i++
			}
		}

		if i > race.MaxChoices {
			invalidRaces = append(invalidRaces, raceID)
		}
	}
	return invalidRaces
}

// Add will add the votes from a ballot to a tally
func (tally *Tally) Add(ballot *Ballot) error {
	for raceID, race := range ballot.Races {
		if tally.raceResults[raceID] == nil {
			tally.raceResults[raceID] = &RaceResult{
				make(map[string]int),
			}
		}
		chosenCandidates := make([]string, 0)
		i := 0
		for candidateName, chosen := range race.Choices {
			if chosen {
				chosenCandidates = append(chosenCandidates, candidateName)
				i++
			}
		}

		if i > race.MaxChoices {
			return common.NewInvalidError("Invalid ballot")
		}

		if len(chosenCandidates) > 0 {
			for _, chosenCandidate := range chosenCandidates {
				tally.raceResults[raceID].results[chosenCandidate]++
			}
		}
	}
	return nil
}

func (tally *Tally) String() string {
	s := ""
	for raceID, raceResult := range tally.raceResults  {
		s += fmt.Sprintf("Race %s:\n", raceID)
		for candidate, count := range raceResult.results {
			s += fmt.Sprintf("\t%s: %d\n", candidate, count)
		}
	}
	return s
}


type Receipt struct {
	voteDigest []byte
	ledgerUuid gUuid.UUID
}

type VotePayload struct {
	Timestamp int64
	VoterElectionUuid gUuid.UUID
	Ballot *Ballot
	BallotSignatureBytes []byte 				// Signed by voter
}

type Ledger struct {
	streamStore entwine.StreamStore
	objectStore storage.ObjectStore
	generatorAuthenticator crypto.Authenticator
	voterAuthenticators map[gUuid.UUID]crypto.Authenticator
	voterElectionUuidToStreamUuid map[gUuid.UUID]gUuid.UUID
	signer crypto.Signer
	authenticator crypto.Authenticator
	municipality entwine.SubStreamID
	currTickerAnchorUuid gUuid.UUID
	lastAnchorHeadUuid gUuid.UUID
}

func NewLedger(municipality entwine.SubStreamID, anchorUuid gUuid.UUID, streamStore entwine.StreamStore,
	objectStore storage.ObjectStore, signer crypto.Signer, authenticator crypto.Authenticator) (*Ledger, error) {

	lastAnchorHead, err := streamStore.Head(municipality)
	if err != nil {
		return nil, err
	}
	lastAnchorHeadUuid := lastAnchorHead.Uuid()

	return &Ledger {
		streamStore: streamStore,
		objectStore: objectStore,
		generatorAuthenticator: nil,
		voterElectionUuidToStreamUuid: make(map[gUuid.UUID]gUuid.UUID),
		voterAuthenticators: make(map[gUuid.UUID]crypto.Authenticator),
		signer: signer,
		authenticator: authenticator,
		municipality: municipality,
		lastAnchorHeadUuid: lastAnchorHeadUuid,
		currTickerAnchorUuid: anchorUuid,
	}, nil
}

func (ledger *Ledger) Municipality() entwine.SubStreamID {
	return ledger.municipality
}

func (ledger *Ledger) SetAnchorUuid(anchorUuid gUuid.UUID) {
	ledger.currTickerAnchorUuid = anchorUuid
}

func (ledger *Ledger) Entwine(ticker entwine.TickerStore) error {
	endNode, err := ledger.streamStore.Head(ledger.municipality)
	if err != nil {
		return err
	}

	endUuid := endNode.Uuid()

	startUuid, err := ticker.GetProofStartUuid(ledger.municipality)
	if err != nil {
		return err
	}

	messages, err := ledger.streamStore.GetHistory(startUuid, endUuid)
	if err != nil {
		return err
	}

	if len(messages) > 0 && strings.Compare(ledger.lastAnchorHeadUuid.String(), endUuid.String()) != 0 {
		anchor, err := ticker.Anchor(messages, ledger.municipality, ledger.authenticator)
		if err != nil {
			return err
		}

		ledger.currTickerAnchorUuid = anchor.Uuid()
	}

	return nil
}

func (ledger *Ledger) GeneratorAuthenticator(generatorAuthenticator crypto.Authenticator) error {
	ledger.generatorAuthenticator = generatorAuthenticator
	return nil
}

func (ledger *Ledger) VoterAuthenticator(voterElectionUuid gUuid.UUID, authenticator crypto.Authenticator) error {
	ledger.voterAuthenticators[voterElectionUuid] = authenticator
	return nil
}

func (ledger *Ledger) Vote(voterElectionUuid gUuid.UUID, voterElectionUuidSignatureBytes []byte, ballotBytes []byte,
	ballotSignatureBytes []byte, testAuth crypto.Authenticator) (*Receipt, error) {

	ballot, err := DeerializeBallot(ballotBytes)
	if err != nil {
		return nil, err
	}
	// Make sure ballot is valid
	invalidRaces := ballot.InvalidRaces()
	if len(invalidRaces) > 0 {
		return nil, common.NewInvalidError(fmt.Sprintf("Invalid choices on ballot: %v", invalidRaces))
	}

	// Check that election UUID is legit
	electionUuidBytes, err := voterElectionUuid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	voterElectionUuidSignature, err := crypto.DeserializeSignature(voterElectionUuidSignatureBytes)
	if err != nil {
		return nil, err
	}
	if !ledger.generatorAuthenticator.Verify(electionUuidBytes, voterElectionUuidSignature) {
		return nil, common.NewInvalidError(fmt.Sprintf("Vote - signature does not match election UUID"))
	}

	// Check that ballot signature is legit
	ballotSignature, err := crypto.DeserializeSignature(ballotSignatureBytes)
	if err != nil {
		return nil, err
	}
	if voterAuthenticator, ok := ledger.voterAuthenticators[voterElectionUuid]; ok {
		if !voterAuthenticator.Verify(ballotBytes, ballotSignature) {
			return nil, common.NewInvalidError(fmt.Sprintf("Vote - signature does not match ballot"))
		}
	}

	byteBuffer := bytes.NewBuffer([]byte{})
	gEncoder := gob.NewEncoder(byteBuffer)
	votePayload := &VotePayload{
		Timestamp: time.Now().Unix(),
		VoterElectionUuid: voterElectionUuid,
		Ballot: ballot,
		BallotSignatureBytes: ballotSignatureBytes,
	}

	err = gEncoder.Encode(&votePayload)
	if err != nil {
		return nil, err
	}

	err = ledger.objectStore.Put(voterElectionUuid.String(), byteBuffer)
	if err != nil {
		return nil, err
	}

	objectParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "default")
	if err != nil {
		return nil, err
	}

	desc := storage.NewObjectDescriptor(objectParams, voterElectionUuid.String())

	message := entwine.NewUncommittedMessage(desc, voterElectionUuid.String(), []string{}, ledger.signer)

	streamUuid, err := ledger.streamStore.Append(message, ledger.municipality, ledger.currTickerAnchorUuid)
	if err != nil {
		return nil, err
	}

	ledger.voterElectionUuidToStreamUuid[voterElectionUuid] = streamUuid

	committedMessage, err := ledger.streamStore.GetMessageByUUID(streamUuid)
	if err != nil {
		return nil, err
	}

	return &Receipt{
		committedMessage.Digest(),
		streamUuid,
	}, nil
}

func MessageToVotePayload(message *entwine.StreamImmutableMessage) (*VotePayload, error) {
	votePayload := &VotePayload{}

	reader, err := message.Data()
	if err != nil {
		return nil, err
	}
	votePayloadBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	byteBuffer := bytes.NewBuffer(votePayloadBytes)
	gDecooder := gob.NewDecoder(byteBuffer)
	err = gDecooder.Decode(&votePayload)
	if err != nil {
		return nil, err
	}

	return votePayload, nil
}

func (ledger *Ledger) Tally() (*Tally, error) {
	currNode, err := ledger.streamStore.Head(ledger.municipality)
	if err != nil {
		return nil, err
	}

	tally := &Tally {
		make(map[RaceID]*RaceResult),
	}

	for {
		if currNode.Prev() == gUuid.Nil {
			break
		}

		votePayload, err := MessageToVotePayload(currNode)
		if err != nil {
			return nil, err
		}

		err = tally.Add(votePayload.Ballot)
		if err != nil {
			return nil, err
		}

		currNode, err = ledger.streamStore.GetMessageByUUID(currNode.Prev())
		if err != nil {
			return nil, err
		}
	}
	return tally, nil
}
