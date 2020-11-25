package main

import (
	gocrypto "crypto"
	"fmt"
	"github.com/kmgreen2/agglo/internal/usecase/voting"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/test"
	"math/rand"
	"sync"
	"time"
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
			ID:     fmt.Sprintf("%d", i),
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
	voters := make([]*voting.Voter, len(people))
	var returnVoters []*voting.Voter
	results := make([]error, len(people))

	wg := &sync.WaitGroup{}
	wg.Add(len(people))

	for i, person := range people {
		if rand.Float64() > probOfRegister {
			wg.Done()
			continue
		}
		go func(index int, person *voting.Person) {
			defer wg.Done()
			voterID, err := registrar.ElectionRegister(person.ID, person.Secret)
			if err != nil {
				results[index] = err
			}
			signer, authenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
			if err != nil {
				results[index] = err
			}
			electionUuid, err := generator.AllocateElectionUuid(voterID, authenticator)
			if err != nil {
				results[index] = err
			}
			voters[index] = voting.NewVoter(person, voterID, electionUuid, signer, authenticator)
		}(i, person)
	}

	wg.Wait()

	for i, _ := range voters {
		if voters[i] != nil {
			returnVoters = append(returnVoters, voters[i])
		}
		if results[i] != nil {
			return nil, results[i]
		}
	}

	return returnVoters, nil
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
	ticker entwine.TickerStore, probOfVoting float64) error {
	for i, voter := range voters {
		if rand.Float64() > probOfVoting {
			continue
		}

		sleepTime := time.Duration(rand.Int() % 50) * time.Millisecond
		time.Sleep(sleepTime)

		if i % 5 == 0 {
			err := ledger.Entwine(ticker)
			if err != nil {
				return err
			}
		}

		person := voter.Person
		electionUuidSignature, err := registrar.PrepareToVote(person.ID, person.Secret, nil)
		if err != nil {
			return err
		}
		ballot := getBallot()
		ballotBytes, err := voting.SerializeBallot(ballot)
		if err != nil {
			return err
		}
		ballotSignature, err := voter.Signer.Sign(ballotBytes)
		if err != nil {
			return err
		}
		ballotSignatureBytes, err := crypto.SerializeSignature(ballotSignature)
		if err != nil {
			return err
		}
		receipt, err:= ledger.Vote(voter.ElectionUuid, electionUuidSignature, ballotBytes, ballotSignatureBytes,
			voter.Authenticator)
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

func startTickerStore() (entwine.TickerStore, error) {
	kvStore := kvs.NewMemKVStore()
	kvTickerStore := entwine.NewKVTickerStore(kvStore, common.SHA1)
	signer, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, err
	}

	err = kvTickerStore.Append(signer)
	if err != nil {
		return nil, err
	}

	go func () {
		for {
			time.Sleep(100*time.Millisecond)
			err = kvTickerStore.Append(signer)
			if err != nil {
				fmt.Printf("WARN - Ticker failure: %s", err.Error())
			}
		}
	}()

	return kvTickerStore, nil
}

func runMunicipality(name string, numPeople int, ticker entwine.TickerStore,
	objectStore storage.ObjectStore, streamStore entwine.StreamStore) (*voting.Ledger,
	*voting.Tally, error) {
	people := createPeople(numPeople)
	registrar := voting.NewRegistrar(people)
	genesisProof, err := ticker.CreateGenesisProof(entwine.SubStreamID(name))
	if err != nil {
		return nil, nil, err
	}

	signer, authenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, err
	}

	err = streamStore.Create(entwine.SubStreamID(name), common.SHA1, signer, genesisProof.TickerUuid())
	if err != nil {
		return nil, nil, err
	}

	ledger, err := voting.NewLedger(entwine.SubStreamID(name), genesisProof.TickerUuid(), streamStore, objectStore,
		signer, authenticator)

	if err != nil {
		return nil, nil, err
	}
	generator, err := voting.NewGenerator(ledger, registrar)
	if err != nil {
		return nil, nil, err
	}

	voters, err := peopleRegister(registrar, generator, people, 1.0)
	if err != nil {
		return nil, nil, err
	}

	err = peopleDoVote(voters, registrar, ledger, ticker, 1.0)
	if err != nil {
		return nil, nil, err
	}

	tally, err := ledger.Tally()
	if err != nil {
		return nil, nil, err
	}

	return ledger, tally, nil
}

func main() {
	ticker, err := startTickerStore()
	if err != nil {
		panic(err)
	}

	numMunicipalities := 4
	numPeople := 20

	kvStore := kvs.NewMemKVStore()

	streamStore := entwine.NewKVStreamStore(kvStore, common.SHA1)

	objectStoreInstance := "default"
	storageParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, objectStoreInstance)
	if err != nil {
		panic(err)
	}
	objectStore, err := storage.NewMemObjectStore(storageParams)
	if err != nil {
		panic(err)
	}

	ledgers := make([]*voting.Ledger, numMunicipalities)
	tallies := make([]*voting.Tally, numMunicipalities)
	wg := &sync.WaitGroup{}
	wg.Add(numMunicipalities)

	for i := 0; i < numMunicipalities; i++ {
		go func(index int) {
			var err error
			defer wg.Done()
			ledgers[index], tallies[index], err = runMunicipality(fmt.Sprintf("muni-%d", index), numPeople, ticker,
				objectStore,
				streamStore)
			if err != nil {
				panic(err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numMunicipalities; i++ {
		fmt.Printf("%s\n", tallies[i].String())
	}

	var allProofs []entwine.Proof
	var allMessages []*entwine.StreamImmutableMessage

	for i := 0; i < numMunicipalities; i++ {
		proofs, err := ticker.GetProofs(entwine.SubStreamID(fmt.Sprintf("muni-%d", i)), 0, -1)
		if err != nil {
			panic(err)
		}
		allProofs = append(allProofs, proofs...)
	}

	entwine.SortProofs(allProofs)

	for _, proof := range allProofs {
		fmt.Println(proof.String())
		if ok, _ := proof.IsGenesis(); ok {
			continue
		}
		messages, err := streamStore.GetHistory(proof.StartUuid(), proof.EndUuid())
		if err != nil {
			panic(err)
		}
		// First message in a proof always contains the last message from the previous epoch (anchored set)
		messages = messages[1:]
		allMessages = append(allMessages, messages...)
	}

	for i, _ := range allMessages {
		if i == 0 {
			continue
		}
		lhs := allMessages[i-1]
		rhs := allMessages[i]
		lhsHappenedBefore, err := lhs.HappenedBefore(rhs, ticker)
		if err != nil {
			panic(err)
		}
		rhsHappenedBefore, err := rhs.HappenedBefore(lhs, ticker)
		if err != nil {
			panic(err)
		}

		// lhs <= rhs
		if lhsHappenedBefore || (!lhsHappenedBefore && !rhsHappenedBefore) {
			continue
		}
		fmt.Printf("Message at idx=%d in %s did not happen before idx=%d in %s\n", lhs.Index(), lhs.SubStream(),
			rhs.Index(), rhs.SubStream())
	}
}