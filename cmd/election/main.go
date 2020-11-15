package main

import (
	"bytes"
	gocrypto "crypto"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/internal/usecase/voting"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
	"math/rand"
)

func chooseRace() voting.RaceType {
	return voting.RaceType(rand.Int() % int(voting.RaceOther+1))
}

func chooseSex() voting.SexType {
	return voting.SexType(rand.Int() % int(voting.OtherSex+1))
}

func chooseAge() int {
	return int(rand.NormFloat64() + 43 + 25) // mean 43, std dev 25
}

func chooseSecret() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func createPeople(numPeople int) []*voting.Person {
	people := make([]*voting.Person, numPeople)

	for i := 0; i < numPeople; i++ {
		people[i] = &voting.Person {
			ID:     fmt.Sprintf("%s", i),
			Secret: chooseSecret(),
			PublicInfo: &voting.PublicInfo{Age: chooseAge(),
				Sex:  chooseSex(),
				Race: chooseRace(),
			},

		}
	}
	return people
}

func peopleRegister(registrar *voting.Registrar, generator *voting.Generator, people []*voting.Person,
	probOfRegister float64) ([]*voting.Voter, error) {
	var voters []*voting.Voter

	for _, person := range people {
		if rand.Float64() > probOfRegister {
			continue
		}
		voterID, err := registrar.ElectionRegister(person.ID, person.Secret)
		if err != nil {
			return nil, err
		}
		signer, authenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
		if err != nil {
			return nil, err
		}
		electionUuid, err := generator.AllocateElectionUuid(voterID, authenticator)
		if err != nil {
			return nil, err
		}
		voters = append(voters, voting.NewVoter(person, voterID, electionUuid, signer, authenticator))
	}
	return voters, nil
}

func getBallot() *voting.Ballot {
	choices := []string{"fizz", "foo", "bar", "buzz"}
	testRace := voting.NewRace(choices, 1)

	randIdx := rand.Int() % len(choices)
	testRace.Choices[choices[randIdx]] = true
	return &voting.Ballot {
		Races: map[voting.RaceID]*voting.Race{
			voting.RaceID("foobar"): testRace,
		},
	}
}

func peopleDoVote(voters []*voting.Voter, registrar *voting.Registrar, ledger *voting.Ledger,
	probOfVoting float64) error {
	for _, voter := range voters {
		if rand.Float64() > probOfVoting {
			continue
		}

		person := voter.Person
		electionUuidSignature, err := registrar.PrepareToVote(person.ID, person.Secret, nil)
		if err != nil {
			return err
		}
		ballot := getBallot()
		byteBuffer := bytes.NewBuffer([]byte{})
		gEncoder := gob.NewEncoder(byteBuffer)
		err = gEncoder.Encode(&ballot)
		if err != nil {
			return err
		}

		ballotSignature, err := voter.Signer.Sign(byteBuffer.Bytes())
		receipt, err:= ledger.Vote(voter.ElectionUuid, electionUuidSignature, ballot, ballotSignature.Bytes())
		if err != nil {
			return err
		}
		err = voter.SetReceipt(receipt)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	people := createPeople(100)
	registrar := voting.NewRegistrar(people)
	ledger, err := voting.NewLedger(entwine.SubStreamID("foobarbaz"))
	if err != nil {
		panic(err.Error())
	}
	generator, err := voting.NewGenerator(ledger, registrar)
	if err != nil {
		panic(err.Error())
	}

	voters, err := peopleRegister(registrar, generator, people, 0.99)
	if err != nil {
		panic(err.Error())
	}

	err = peopleDoVote(voters, registrar, ledger, 0.95)
	if err != nil {
		panic(err.Error())
	}

}
