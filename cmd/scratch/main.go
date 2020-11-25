package main

import (
	"fmt"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/util/random"
)

func main(){
	msg := []byte("Hello Boneh-Lynn-Shacham")
	msg2 := []byte("ello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private1, public1 := bls.NewKeyPair(suite, random.New())
	private2, public2 := bls.NewKeyPair(suite, random.New())
	sig1, err := bls.Sign(suite, private1, msg)
	if err != nil {
		panic(err.Error())
	}
	sig2, err := bls.Sign(suite, private2, msg2)
	if err != nil {
		panic(err.Error())
	}
	aggregatedSig, err := bls.AggregateSignatures(suite, sig1, sig2)
	if err != nil {
		panic(err.Error())
	}

	aggregatedKey := bls.AggregatePublicKeys(suite, public1, public2)

	err = bls.BatchVerify(suite, []kyber.Point{public1, public2}, [][]byte{msg, msg2}, aggregatedSig)
	if err != nil {
		panic(err.Error())
	}

	err = bls.Verify(suite, aggregatedKey, msg, aggregatedSig)
	if err == nil {
		fmt.Printf("Should not be valid")
	}
}
