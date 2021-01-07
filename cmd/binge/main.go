package main

import (
	"flag"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/serialization"
	"github.com/kmgreen2/agglo/pkg/server"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type RunType int

const (
	RunStandalone RunType = iota
	RunDaemon
)

type CommandArgs struct {
	config *os.File
	outfile *os.File
	runType RunType
	daemonPort int
	daemonPath string
	maxConnections int
	numThreads int
	force bool
	stateDbPath string
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() *CommandArgs {
	var err error
	args := &CommandArgs{}
	configPtr := flag.String("config", "", "path to config file for binge")
	outfilePtr := flag.String("outfile", "/dev/stdout", "path to file to store output")
	runTypePtr := flag.String("runType", "standalone",
		"run type for the process: standalone (default) or daemon")
	daemonPortPtr := flag.Int("daemonPort", 8080, "daemon listening port (default 8080)")
	daemonPathPtr := flag.String("daemonPath", "/binge", "daemon processing path (default /binge)")
	maxConnectionsPtr := flag.Int("maxConnections", 64, "maximum number of connections")
	numThreadsPtr := flag.Int("numThreads", 16, "number of worker threads")
	stateDbPathPtr := flag.String("stateDbPath", "/tmp/bingeDb", "location of the state db file")
	forcePtr := flag.Bool("force", false, "force overwrite state entries")

	flag.Parse()

	if strings.Compare(*runTypePtr, "standalone") == 0 {
		args.runType = RunStandalone
	} else if strings.Compare(*runTypePtr, "daemon") == 0 {
		args.runType = RunDaemon
	} else {
		panic(fmt.Sprintf("invalid runType '%s'", *runTypePtr))
	}

	args.maxConnections = *maxConnectionsPtr
	args.numThreads = *numThreadsPtr
	args.force = *forcePtr
	args.stateDbPath = *stateDbPathPtr

	if len(*configPtr) == 0 {
		usage("must specify -config", 1)
	} else {
		args.config, err = os.Open(*configPtr)
		if err != nil {
			usage(err.Error(), 1)
		}
	}

	if args.runType == RunStandalone {
		if len(*outfilePtr) == 0 {
			usage("must specify -outfile", 1)
		} else {
			args.outfile, err = os.OpenFile(*outfilePtr, os.O_RDWR | os.O_CREATE, os.ModePerm)
			if err != nil {
				usage(err.Error(), 1)
			}
		}
	} else if args.runType == RunDaemon {
		args.daemonPath = *daemonPathPtr
		args.daemonPort = *daemonPortPtr
	}

	return args
}

func main() {
	args := parseArgs()

	configBytes, err := ioutil.ReadAll(args.config)
	if err != nil {
		panic(err)
	}

	pipelines, err := serialization.PipelinesFromJson(configBytes)
	if err != nil {
		panic(err)
	}

	// RunStandalone will invoke each pipeline synchronously
	if args.runType == RunStandalone {
		var in map[string]interface{}
		payloadBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		in, err = common.JsonToMap(payloadBytes)
		if err != nil {
			panic(err)
		}

		for _, pipeline := range pipelines {
			out, err := pipeline.RunSync(in)
			if err != nil {
				panic(err)
			}
			outBytes, err := common.MapToJson(out)

			_, err = io.WriteString(args.outfile, string(outBytes))
			if err != nil {
				panic(err)
			}
		}
	} else if args.runType == RunDaemon {
		recoverFunc := func(inBytes []byte) error {
			in, err := common.JsonToMap(inBytes)
			if err != nil {
				return nil
			}
			return server.RunPipelines(in, pipelines)
		}

		dQueue, err := common.OpenDurableQueue(args.stateDbPath, recoverFunc, args.force)
		if err != nil {
			panic(err)
		}

		daemon, err := server.NewDaemon(args.daemonPort, args.daemonPath, args.maxConnections, args.numThreads,
			dQueue, pipelines)
		if err != nil {
			panic(err)
		}
		err = daemon.Run()
		if err != nil {
			panic(err)
		}
	}
}