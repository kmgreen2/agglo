package voting

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/crypto"
)

type Generator struct {
	electionUuidToVoterID map[gUuid.UUID]gUuid.UUID
	electionUuidToAuthenticator map[gUuid.UUID]crypto.Authenticator
	voterIDToElectionUuid map[gUuid.UUID]gUuid.UUID
	signer crypto.Signer
	authenticator crypto.Authenticator
	ledgerBackChannel LedgerBackChannel
	registrarBackChannel RegistrarBackChannel
}

// Allocate an uuid for this election
func (generator *Generator) AllocateElectionUuid(voterID gUuid.UUID,
	voterAuthenticator crypto.Authenticator) (gUuid.UUID, error) {
	if electionUuid, ok := generator.voterIDToElectionUuid[voterID]; ok {
		return electionUuid, nil
	}
	electionUuid := gUuid.New()
	generator.voterIDToElectionUuid[voterID] = electionUuid
	generator.electionUuidToAuthenticator[voterID] = voterAuthenticator

	err := generator.ledgerBackChannel.VoterAuthenticator(electionUuid, voterAuthenticator)
	if err != nil {
		return electionUuid, err
	}

	return electionUuid, nil
}
