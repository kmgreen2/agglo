package server

import (
	"context"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/kmgreen2/agglo/pkg/common"
	"golang.org/x/net/netutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type Daemon interface {
	Run() error
	Shutdown() error
	IsShutdown() bool
}

func RunPipelines(in map[string]interface{}, pipelines *process.Pipelines) error {
	for _, pipeline := range pipelines.Underlying() {
		future := pipeline.RunAsync(in)
		result := future.Get()
		if result.Error() != nil {
			return result.Error()
		}
	}
	return nil
}

type StatelessDaemon struct {
	srv *http.Server
	pipelines *process.Pipelines
	logger *zap.Logger
	daemonPort int
	daemonPath string
	maxConnections int
	shutdown bool
	isRunning bool
}

func NewStatelessDaemon(daemonPort int, daemonPath string, maxConnections int,
	pipelines *process.Pipelines) (Daemon, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &StatelessDaemon{
		pipelines: pipelines,
		logger: logger,
		daemonPath: daemonPath,
		daemonPort: daemonPort,
		maxConnections: maxConnections,
	}, nil
}

func (d *StatelessDaemon) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", d.daemonPort))
	if err != nil {
		return err
	}

	defer l.Close()

	l = netutil.LimitListener(l, d.maxConnections)

	d.srv = &http.Server{}

	http.HandleFunc(d.daemonPath, func(resp http.ResponseWriter, req *http.Request) {
		begin := time.Now()
		defer func() {
			elapsed := time.Now().Sub(begin)
			d.logger.Info(fmt.Sprintf("Http handler: %d ms", elapsed.Milliseconds()))
		}()
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			resp.WriteHeader(400)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}
		in, err := common.JsonToMap(bodyBytes)
		if err != nil {
			resp.WriteHeader(400)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		// A globally unique ID is defined which will allow the re-player to fetch the
		// checkpoint state, prior to replaying, among other future uses.
		// It is also assumed that each pipeline mix in its own name to ensure the
		// there are no conflicts with checkpoint state
		messageID, err := gUuid.NewRandom()
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		in["agglo:messageID"] = messageID.String()
		err = RunPipelines(in, d.pipelines)
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		resp.WriteHeader(200)
		_, _ = resp.Write([]byte("Success"))
		return
	})

	d.isRunning = true
	return d.srv.Serve(l)
}

func (d *StatelessDaemon) IsShutdown() bool {
	return d.shutdown && !d.isRunning
}

func (d *StatelessDaemon) Shutdown() error {
	d.shutdown = true
	defer func(){d.isRunning = false}()
	d.logger.Info("shutting down daemon...")
	if err := d.srv.Shutdown(context.Background()); err != nil {
		d.logger.Error(err.Error())
	}
	if err := d.pipelines.Shutdown(); err != nil {
		d.logger.Error(err.Error())
	}
	d.logger.Info("shutdown complete...")
	return nil
}

type DurableDaemon struct {
	srv *http.Server
	pipelines *process.Pipelines
	logger *zap.Logger
	daemonPort int
	daemonPath string
	maxConnections int
	numThreads int
	dQueue *common.DurableQueue
	waiterChannel chan bool
	workerSleeping bool
	shutdown bool
	isRunning bool
}

func NewDurableDaemon(daemonPort int, daemonPath string, maxConnections, numThreads int,
	dQueue *common.DurableQueue, pipelines *process.Pipelines) (Daemon, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &DurableDaemon{
		pipelines: pipelines,
		logger: logger,
		daemonPath: daemonPath,
		daemonPort: daemonPort,
		numThreads: numThreads,
		maxConnections: maxConnections,
		dQueue: dQueue,
		waiterChannel: make(chan bool, 1),
	}, nil
}

func (d *DurableDaemon) startWorkerLoop() {
	go func() {
		threadChannel := make(chan bool, d.numThreads)
		for {

			if d.shutdown == true {
				break
			}

			item, err := d.dQueue.Dequeue()
			if err != nil {
				// If the queue is empty, go to sleep until the producer queues up more messages
				if errors.Is(err, &common.EmptyQueue{}) {
					d.workerSleeping = true
					<- d.waiterChannel
					d.workerSleeping = false
					continue
				} else {
					d.logger.Error(err.Error())
				}
			}

			begin := time.Now()
			threadChannel <- true
			go func(item *common.QueueItem) {
				in, err := common.JsonToMap(item.Data)
				if err != nil {
					d.logger.Error("unable to decode data: " + err.Error())
				}
				err = RunPipelines(in, d.pipelines)
				if err != nil {
					d.logger.Error("error running pipeline: " + err.Error())
				}
				err = d.dQueue.Ack(item)
				if err != nil {
					d.logger.Error("error acking item: " + err.Error())
				}
				elapsed := time.Now().Sub(begin)
				d.logger.Info(fmt.Sprintf("Worker: %d ms", elapsed.Milliseconds()))
				<- threadChannel
			}(item)
		}
	}()
}

func (d *DurableDaemon) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", d.daemonPort))
	if err != nil {
		return err
	}

	defer l.Close()

	l = netutil.LimitListener(l, d.maxConnections)

	d.srv = &http.Server{}

	d.startWorkerLoop()

	http.HandleFunc(d.daemonPath, func(resp http.ResponseWriter, req *http.Request) {
		begin := time.Now()
		defer func() {
			elapsed := time.Now().Sub(begin)
			d.logger.Info(fmt.Sprintf("Http handler: %d ms", elapsed.Milliseconds()))
		}()
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			resp.WriteHeader(400)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}
		in, err := common.JsonToMap(bodyBytes)
		if err != nil {
			resp.WriteHeader(400)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		// A globally unique ID is defined which will allow the re-player to fetch the
		// checkpoint state, prior to replaying, among other future uses.
		// It is also assumed that each pipeline mix in its own name to ensure the
		// there are no conflicts with checkpoint state
		messageID, err := gUuid.NewRandom()
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		in["agglo:messageID"] = messageID.String()
		inBytes, err := common.MapToJson(in)
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		err = d.dQueue.Enqueue(inBytes)
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		// If the worker thread is sleeping, wake it up
		if d.workerSleeping {
			d.waiterChannel <- true
		}

		resp.WriteHeader(200)
		_, _ = resp.Write([]byte("Success"))
		return
	})

	d.isRunning = true
	return d.srv.Serve(l)
}

func (d *DurableDaemon) IsShutdown() bool {
	return d.shutdown && !d.isRunning
}

func (d *DurableDaemon) Shutdown() error {
	d.shutdown = true
	defer func(){d.isRunning = false}()
	d.logger.Info("shutting down daemon...")
	if err := d.srv.Shutdown(context.Background()); err != nil {
		d.logger.Error(err.Error())
	}
	if err := d.dQueue.Close(); err != nil {
		d.logger.Error(err.Error())
	}
	if err := d.pipelines.Shutdown(); err != nil {
		d.logger.Error(err.Error())
	}
	d.logger.Info("shutdown complete...")
	return nil
}
