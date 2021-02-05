package voting

import (
	gocrypto "crypto"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/test"
	"sync"
)

type Generator struct {
	electionUuidToVoterID map[gUuid.UUID]gUuid.UUID
	electionUuidToAuthenticator map[gUuid.UUID]crypto.Authenticator
	voterIDToElectionUuid map[gUuid.UUID]gUuid.UUID
	signer crypto.Signer
	authenticator crypto.Authenticator
	ledgerBackChannel LedgerBackChannel
	registrarBackChannel RegistrarBackChannel
	lock *sync.Mutex
}

func NewGenerator(ledgerBackChannel LedgerBackChannel, registrarBackChannel RegistrarBackChannel) (*Generator, error) {
	signer, authenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, err
	}
	generator := &Generator{
		electionUuidToVoterID: make(map[gUuid.UUID]gUuid.UUID),
		electionUuidToAuthenticator: make(map[gUuid.UUID]crypto.Authenticator),
		voterIDToElectionUuid: make(map[gUuid.UUID]gUuid.UUID),
		signer: signer,
		authenticator: authenticator,
		ledgerBackChannel: ledgerBackChannel,
		registrarBackChannel: registrarBackChannel,
		lock: &sync.Mutex{},
	}

	err = ledgerBackChannel.GeneratorAuthenticator(authenticator)
	if err != nil {
		return nil, err
	}

	return generator, nil
}

// ListElectionUuids will list all of the registered UUIDs
func (generator *Generator) ListElectionUuids() ([]gUuid.UUID, error) {
	allUuids := make([]gUuid.UUID, 0)
	for electionUuid, _ := range generator.electionUuidToVoterID {
		allUuids = append(allUuids, electionUuid)
	}
	return allUuids, nil
}

// Allocate an uuid for this election
// Note: This needs to be idempotent!
func (generator *Generator) AllocateElectionUuid(voterID gUuid.UUID,
	voterAuthenticator crypto.Authenticator) (gUuid.UUID, error) {
	generator.lock.Lock()
	defer generator.lock.Unlock()
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

	electionUuidBytes, err := electionUuid.MarshalBinary()
	if err != nil {
		return electionUuid, err
	}

	signature, err := generator.signer.Sign(electionUuidBytes)
	if err != nil {
		return electionUuid, nil
	}

	signatureBytes, err := crypto.SerializeSignature(signature)
	if err != nil {
		return electionUuid, nil
	}

	err = generator.registrarBackChannel.ElectionUUIDCommit(voterID, signatureBytes)

	return electionUuid, err
}
