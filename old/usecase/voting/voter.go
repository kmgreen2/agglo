package voting

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/crypto"
)

type SexType int
const (
	Male SexType = iota
	Female
	Nonbinary
	OtherSex
)

type RaceType int
const (
	RaceCaucasian RaceType = iota
	RaceAfricanAmerican
	RaceNativeAmerican
	RaceHispanic
	RaceAsian
	RaceOther
)

type PublicInfo struct {
	Age int
	Sex SexType
	Race RaceType
}

type Person struct {
	ID     string
	Secret string // Obtained via QR code: either in-person by showing ID or remotely by scanning ID
				  // This is needed to get a voterID
	PublicInfo *PublicInfo
}

type Voter struct {
	Person *Person
	VoterID gUuid.UUID
	ElectionUuid gUuid.UUID
	receipt *Receipt
	Signer crypto.Signer
	Authenticator crypto.Authenticator
}

func NewVoter(person *Person, voterID gUuid.UUID, electionUuid gUuid.UUID, signer crypto.Signer,
	authenticator crypto.Authenticator) *Voter {
	return &Voter {
		Person: person,
		VoterID: voterID,
		ElectionUuid: electionUuid,
		receipt: nil,
		Signer: signer,
		Authenticator: authenticator,
	}
}

func (voter *Voter) SetReceipt(receipt *Receipt) error {
	voter.receipt = receipt
	return nil
}

