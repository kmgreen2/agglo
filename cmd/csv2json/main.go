package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"encoding/csv"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type CommandArgs struct {
	csvFile *os.File
	outfile *os.File
	outEndpoint string
	csvMap map[int64]string
	fieldMatchers map[int64]*regexp.Regexp
	debugLogging bool
	concurrencyChan chan bool
	outfileMutex sync.Mutex
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() *CommandArgs {
	var err error
	args := &CommandArgs{}
	args.csvMap = make(map[int64]string)
	args.fieldMatchers = make(map[int64]*regexp.Regexp)
	csvFilePtr := flag.String("csvFile", "/dev/stdin", "path to fetch CSV input")
	outfilePtr := flag.String("outfile", "", "path to file to store output")
	outEndpointPtr := flag.String("outEndpoint", "", "path to endpoint to send output")
	csvMapPtr := flag.String("csvMap", "",
		"Comma-delimited mapping of CSV index to output field name: idx1:field1,idx2:field2,...")
	debugLoggingPtr := flag.Bool("debugLogging", false, "enable debug logging")
	concurrencyPtr := flag.Int("concurrency", 1, "concurrency")
	fieldMatchers := flag.String("fieldMatchers", "",
		"Comma-delimited field to regex matchers: idx1:r1,idx2:r2,...")

	flag.Parse()

	args.csvFile, err = os.Open(*csvFilePtr)
	if err != nil {
		panic(err)
	}

	args.debugLogging = *debugLoggingPtr
	args.concurrencyChan = make(chan bool, *concurrencyPtr)

	if len(*outfilePtr) == 0 && len(*outEndpointPtr) == 0 {
		panic("Must specify either output file (-outfile) or output endpoint (-outEndpoint)")
	}
	if len(*outfilePtr) > 0 && len(*outEndpointPtr) > 0 {
		panic("Cannot specify both output file (-outfile) and output endpoint (-outEndpoint)")
	}

	if len(*outfilePtr) != 0 {
		args.outfile, err = os.Open(*outfilePtr)
		if err != nil {
			panic(err)
		}
	} else {
		args.outEndpoint = *outEndpointPtr
	}

	if len(*csvMapPtr) == 0 {
		panic("Missing CSV mapping")
	}

	for _, m := range strings.Split(*csvMapPtr, ",") {
		mAry := strings.Split(m, ":")
		if len(mAry) != 2 {
			panic(fmt.Sprintf("Malformed CSV mapping entry: %s.  Expected idx:field", m))
		}
		idx, err := strconv.ParseInt(mAry[0], 10, 32)
		if err != nil {
			panic(err)
		}
		args.csvMap[idx] = mAry[1]
	}

	for _, m := range strings.Split(*fieldMatchers, ",") {
		mAry := strings.Split(m, ":")
		if len(mAry) != 2 {
			panic(fmt.Sprintf("Malformed field matcher entry: %s.  Expected idx:regex", m))
		}
		idx, err := strconv.ParseInt(mAry[0], 10, 32)
		if err != nil {
			panic(err)
		}
		cRegex, err := regexp.Compile(mAry[1])
		if err != nil {
			panic(err)
		}
		args.fieldMatchers[idx] = cRegex
	}


	return args
}

func sendOutput(args *CommandArgs, jsonBuf *bytes.Buffer) {
	if len(args.outEndpoint) > 0 {
		resp, err := http.Post(args.outEndpoint, "application/json", jsonBuf)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			panic(fmt.Sprintf("Error posting JSON to %s: %s", args.outEndpoint, resp.Status))
		} else if args.debugLogging {
			fmt.Printf("Status:%d\n", resp.StatusCode)
		}
		_ = resp.Body.Close()
	} else {
		numBytes := len(jsonBuf.Bytes())
		args.outfileMutex.Lock()
		for numBytes > 0 {
			n, err := args.outfile.Write(jsonBuf.Bytes())
			if err != nil {
				panic(err)
			}
			numBytes -= n
		}
		args.outfileMutex.Unlock()
	}
	<- args.concurrencyChan
}

func main() {
	args := parseArgs()

	csvReader := csv.NewReader(args.csvFile)

	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, record := range records {
		outMap := make(map[string]string)
		match := true
		for idx, field := range args.csvMap {
			if cRegex, ok := args.fieldMatchers[idx]; ok {
				if !cRegex.MatchString(record[idx]) {
					match = false
					break
				}
			}
			outMap[field] = record[idx]
		}
		if !match {
			continue
		}
		jsonBuf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(jsonBuf)
		err = encoder.Encode(outMap)
		if err != nil {
			panic(err)
		}
		args.concurrencyChan <- true
		go sendOutput(args, jsonBuf)
	}
}
