package main

import (
	"flag"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/observability"
	"github.com/kmgreen2/agglo/pkg/serialization"
	"github.com/kmgreen2/agglo/pkg/server"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
	"sync"
)

type RunType int

const (
	RunStandalone RunType = iota
	RunStatelessDaemon
	RunPersistentDaemon
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
	exporter observability.Exporter
	profileStopFn func()
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func registerSignalHandler(daemon server.Daemon) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func () {
		<-c
		wg.Add(1)
		_ = daemon.Shutdown()
		wg.Done()
	}()
	return wg
}

func parseArgs() *CommandArgs {
	var err error
	args := &CommandArgs{}
	configPtr := flag.String("config", "", "path to config file for binge")
	outfilePtr := flag.String("outfile", "/dev/stdout", "path to file to store output")
	runTypePtr := flag.String("runType", "standalone",
		"run type for the process: standalone (default), stateless-daemon, persistent-daemon")
	daemonPortPtr := flag.Int("daemonPort", 8080, "daemon listening port (default 8080)")
	daemonPathPtr := flag.String("daemonPath", "/binge", "daemon processing path (default /binge)")
	maxConnectionsPtr := flag.Int("maxConnections", 64, "maximum number of connections")
	numThreadsPtr := flag.Int("numThreads", 16, "number of worker threads")
	stateDbPathPtr := flag.String("stateDbPath", "/tmp/bingeDb", "location of the state db file")
	exporterPtr := flag.String("exporter", "stdout", "OpenTelemetry exporter type")
	forcePtr := flag.Bool("force", false, "force overwrite state entries")
	cpuProfilePtr := flag.String("cpuprofile", "", "write the CPU profile to a file")

	flag.Parse()

	args.profileStopFn = func() {
	}

	if cpuProfilePtr != nil {
		if *cpuProfilePtr != "" {
			f, err := os.Create(*cpuProfilePtr)
			if err != nil {
				usage("must specify file location with cpuprofile", 1)
			}
			err = pprof.StartCPUProfile(f)
			if err != nil {
				usage(err.Error(), 1)
			}
			args.profileStopFn = func() {
				pprof.StopCPUProfile()
			}
		}
	}

	if strings.Compare(*runTypePtr, "standalone") == 0 {
		args.runType = RunStandalone
	} else if strings.Compare(*runTypePtr, "persistent-daemon") == 0 {
		args.runType = RunPersistentDaemon
	} else if strings.Compare(*runTypePtr, "stateless-daemon") == 0 {
		args.runType = RunStatelessDaemon
	} else {
		panic(fmt.Sprintf("invalid runType '%s'", *runTypePtr))
	}

	args.maxConnections = *maxConnectionsPtr
	args.numThreads = *numThreadsPtr
	args.force = *forcePtr
	args.stateDbPath = *stateDbPathPtr

	if strings.Compare(*exporterPtr, "stdout") == 0 {
		args.exporter, err = observability.NewStdoutExporter()
		if err != nil {
			panic(err)
		}
	} else if strings.Compare(*exporterPtr, "zipkin") == 0 {
		args.exporter = observability.NewZipkinExporter("http://localhost:9411/api/v2/spans", "binge")
	} else {
		usage(fmt.Sprintf("invalid exporter: %s", *exporterPtr), 1)
	}

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
	} else if args.runType == RunStatelessDaemon || args.runType == RunPersistentDaemon {
		args.daemonPath = *daemonPathPtr
		args.daemonPort = *daemonPortPtr
	}

	return args
}

func main() {
	args := parseArgs()

	defer args.profileStopFn()

	configBytes, err := ioutil.ReadAll(args.config)
	if err != nil {
		panic(err)
	}

	pipelines, err := serialization.PipelinesFromJson(configBytes)
	if err != nil {
		panic(err)
	}

	err = args.exporter.Start()
	if err != nil {
		panic(err)
	}

	defer func(commandArgs *CommandArgs) { _ = args.exporter.Stop() }(args)

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

		for _, pipeline := range pipelines.Underlying() {
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
		_ = pipelines.Shutdown()
	} else if args.runType == RunPersistentDaemon {
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

		daemon, err := server.NewDurableDaemon(args.daemonPort, args.daemonPath, args.maxConnections, args.numThreads,
			dQueue, pipelines)
		if err != nil {
			panic(err)
		}
		waiter := registerSignalHandler(daemon)
		err = daemon.Run()
		waiter.Wait()
		if !daemon.IsShutdown() && err != nil {
			panic(err)
		}
	} else if args.runType == RunStatelessDaemon {
		daemon, err := server.NewStatelessDaemon(args.daemonPort, args.daemonPath, args.maxConnections, pipelines)
		if err != nil {
			panic(err)
		}
		waiter := registerSignalHandler(daemon)
		err = daemon.Run()
		waiter.Wait()
		if !daemon.IsShutdown() && err != nil {
			panic(err)
		}
	}
}