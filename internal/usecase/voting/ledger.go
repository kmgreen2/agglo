package voting

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
)

type LedgerBackChannel interface {
	GeneratorAuthenticator(authenticator crypto.Authenticator) error
	VoterAuthenticator(voterElectionUuid gUuid.UUID, authenticator crypto.Authenticator) error
}

type Ballot struct {
}

type Receipt struct {
	voteDigest []byte
}

type VotePayload struct {
	timestamp int64
	voterElectionUuid gUuid.UUID
	ballot *Ballot
	ballotSignature []byte 				// Signed by voter
	voterElectionUuidSignature []byte  	// Signed by UUID generator
}

type Ledger struct {
	streamStore *entwine.KVStreamStore
	generatorAuthenticator crypto.Authenticator
	voterAuthenticators map[gUuid.UUID]crypto.Authenticator
}

func (ledger *Ledger) Vote(voterElectionUuid gUuid.UUID, voterElectionUuidSignature []byte, ballot *Ballot,
	ballotSignature []byte) (*Receipt, error) {
	return nil, nil
}
