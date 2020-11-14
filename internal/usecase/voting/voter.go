package voting

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/crypto"
)

type Person struct {
	id string
	secret string // Obtained via QR code: either in-person by showing ID or remotely by scanning ID
				  // This is needed to get a voterID
}

type Voter struct {
	person *Person
	voterID gUuid.UUID
	electionUuid gUuid.UUID
	voteDigest []byte
	signer crypto.Signer
	authenticator crypto.Authenticator
}
