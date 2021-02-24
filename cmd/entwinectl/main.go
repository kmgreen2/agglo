package main

import (
	"context"
	"flag"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/client"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/util"
	"google.golang.org/grpc"
	"os"
	"strconv"
	"strings"
)

type EntwineCtlArgs struct {
	tickerClient client.TickerClient
	kvStore      kvs.KVStore
	kvStoreConnectionString string
	commandType CommandType
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

type CommandType string

const (
	Unknown CommandType = "Unknown"
	HappenedBefore = "HappenedBefore"
)

func doHappenedBefore(tickerClient client.TickerClient, args... string) error {
	if len(args) != 2 {
		msg := fmt.Sprintf("expected lhs and rhs specs for substream messages (subStreamID:UUID:streamIdx)")
		return util.NewInvalidError(msg)
	}

	lhsAry := strings.Split(args[0], ":")
	if len(lhsAry) != 3 {
		msg := fmt.Sprintf("lhs spec is invalid, expected <subStreamID>:<UUID>:<streamIdx>")
		return util.NewInvalidError(msg)
	}

	lhsStreamIdx, err := strconv.ParseInt(lhsAry[2], 10, 64)
	if err != nil {
		msg := fmt.Sprintf("lhs spec has invalid streamIdx: %s", err.Error())
		return util.NewInvalidError(msg)
	}

	lhsSubStreamID, lhsUuid := lhsAry[0], gUuid.MustParse(lhsAry[1])

	rhsAry := strings.Split(args[1], ":")
	if len(rhsAry) != 3 {
		msg := fmt.Sprintf("rhs spec is invalid, expected <subStreamID>:<UUID>:<streamIdx>")
		return util.NewInvalidError(msg)
	}

	rhsStreamIdx, err := strconv.ParseInt(rhsAry[2], 10, 64)
	if err != nil {
		msg := fmt.Sprintf("rhs spec has invalid streamIdx: %s", err.Error())
		return util.NewInvalidError(msg)
	}

	rhsSubStreamID, rhsUuid := rhsAry[0], gUuid.MustParse(rhsAry[1])

	result, err := tickerClient.HappenedBefore(context.Background(), lhsSubStreamID, lhsUuid, lhsStreamIdx,
		rhsSubStreamID, rhsUuid, rhsStreamIdx)

	if err != nil {
		fmt.Printf("%s happenedBefore %s: %v\n", lhsUuid.String(), rhsUuid.String(), result)
	}
	return err
}

func parseArgs() (*EntwineCtlArgs, error) {
	var err error
	args := &EntwineCtlArgs{}

	kvConnectionStringPtr := flag.String("kvConnectionString", "",
		"KVStore connection string (default:  mem:local)")
	tickerEndpointPtr := flag.String("tickerEndpoint", "", "host:port formatted gRPC endpoint")
	commandPtr := flag.String("command", "", "")

	flag.Parse()


	if len(*tickerEndpointPtr) > 0 {
		args.tickerClient, err = client.NewTickerClient(*tickerEndpointPtr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
	}

	if len(*kvConnectionStringPtr) > 0 {
		args.kvStoreConnectionString = *kvConnectionStringPtr
	}

	args.commandType = CommandType(*commandPtr)

	switch args.commandType {
	case HappenedBefore:
		if args.tickerClient == nil {
			usage(fmt.Sprintf("tickerEndpoint requied for %s", args.commandType), 2)
		}
		err = doHappenedBefore(args.tickerClient, flag.Args()...)
		if err != nil {
			usage(err.Error(), 2)
		}
	default:
		usage(fmt.Sprintf("unknown command type: %s", args.commandType), 2)
	}

	return args, nil
}

func main() {

	args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	switch args.commandType {
	case HappenedBefore:
		if args.tickerClient == nil {
			usage(fmt.Sprintf("tickerEndpoint requied for %s", args.commandType), 2)
		}
		err = doHappenedBefore(args.tickerClient, flag.Args()...)
		if err != nil {
			usage(err.Error(), 2)
		}
	default:
		usage(fmt.Sprintf("unknown command type: %s", args.commandType), 2)
	}
}

