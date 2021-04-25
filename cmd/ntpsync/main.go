package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beevik/ntp"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/util"
	"io/ioutil"
	"os"
)

type NtpSyncArgs struct {
	server string
}

func parseArgs() (*NtpSyncArgs, error) {
	args := &NtpSyncArgs{}

	serverPtr := flag.String("server", "time.google.com",
		"NTP server to connect to")

	flag.Parse()

	if len(*serverPtr) > 0 {
		args.server = *serverPtr
	}

	return args, nil
}

func main() {
	var inMap, outMap map[string]interface{}

	args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	decodeBuffer := bytes.NewBuffer(inBytes)
	decoder := json.NewDecoder(decodeBuffer)
	err = decoder.Decode(&inMap)
	if err != nil {
		panic(err)
	}

	outMap = util.CopyableMap(inMap).DeepCopy()


	time, err := ntp.Time(args.server)
	if err != nil {
		panic(err)
	}

	outMap[string(common.NTPTimeKey)] = time.Unix()

	encodeBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(encodeBuffer)
	err = encoder.Encode(outMap)
	if err != nil {
		panic(err)
	}
	fmt.Print(encodeBuffer.String())
}