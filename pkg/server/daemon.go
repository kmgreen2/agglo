package server

import (
	"bytes"
	"context"
	"encoding/json"
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
	maintenanceSrv *http.Server
	pipelines *process.Pipelines
	logger *zap.Logger
	daemonPort int
	daemonPath string
	maintenancePort int
	maintenancePath string
	maxConnections int
	shutdown bool
	isRunning bool
}

func NewStatelessDaemon(daemonPort, maintenancePort int, daemonPath, maintenancePath string, maxConnections int,
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
		maintenancePath: maintenancePath,
		maintenancePort: maintenancePort,
		maxConnections: maxConnections,
	}, nil
}

func setMaintenanceServerHandlers(path string, pipelines *process.Pipelines, logger *zap.Logger) {
	checkPointsEndpoint := path + "/checkpoints"
	http.HandleFunc(checkPointsEndpoint, func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" || len(req.Method) == 0 {
			queryParams := req.URL.Query()
			messageID := queryParams.Get("messageId")

			checkpoints, err := pipelines.GetCheckpoints(messageID)
			if err != nil {
				if errors.Is(err, &common.NotFoundError{}) {
					resp.WriteHeader(404)
				} else {
					resp.WriteHeader(500)
				}
				_, _ = resp.Write([]byte(fmt.Sprintf(err.Error())))
				logger.Error(err.Error())
				return
			}
			byteBuffer := bytes.NewBuffer([]byte{})
			encoder := json.NewEncoder(byteBuffer)
			err = encoder.Encode(map[string]interface{}{
				"checkpoints": checkpoints,
			})
			if err != nil {
				resp.WriteHeader(500)
				_, _ = resp.Write([]byte(fmt.Sprintf(err.Error())))
				logger.Error(err.Error())
				return
			}
			resp.Header().Set("context-type", "application/json")
			resp.WriteHeader(200)
			_, _ = resp.Write(byteBuffer.Bytes())
		} else {
			resp.WriteHeader(400)
			msg := fmt.Sprintf("%s does not support '%s'", checkPointsEndpoint, req.Method)
			_, _ = resp.Write([]byte(fmt.Sprintf(msg)))
			logger.Error(msg)
			return
		}
	})
}

func (d *StatelessDaemon) runMaintenanceServer() error {
	d.maintenanceSrv = &http.Server{Addr: fmt.Sprintf(":%d", d.maintenancePort)}

	setMaintenanceServerHandlers(d.maintenancePath, d.pipelines, d.logger)

	return d.maintenanceSrv.ListenAndServe()
}

func (d *StatelessDaemon) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", d.daemonPort))
	if err != nil {
		return err
	}

	defer l.Close()

	l = netutil.LimitListener(l, d.maxConnections)

	go func() {
		_ = d.runMaintenanceServer()
	}()

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
	if err := d.maintenanceSrv.Shutdown(context.Background()); err != nil {
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
	maintenanceSrv *http.Server
	logger *zap.Logger
	daemonPort int
	daemonPath string
	maintenancePort int
	maintenancePath string
	maxConnections int
	numThreads int
	dQueue *common.DurableQueue
	waiterChannel chan bool
	workerSleeping bool
	shutdown bool
	isRunning bool
}

func NewDurableDaemon(daemonPort, maintenancePort int, daemonPath, maintenancePath string, maxConnections,
	numThreads int,
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
		maintenancePath: maintenancePath,
		maintenancePort: maintenancePort,
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
				defer func() {
					<- threadChannel
				}()
				in, err := common.JsonToMap(item.Data)
				if err != nil {
					d.logger.Error("unable to decode data: " + err.Error())
					return
				}
				err = RunPipelines(in, d.pipelines)
				if err != nil {
					d.logger.Error("error running pipeline: " + err.Error())
					return
				}
				err = d.dQueue.Ack(item)
				if err != nil {
					d.logger.Error("error acking item: " + err.Error())
					return
				}
				elapsed := time.Now().Sub(begin)
				d.logger.Info(fmt.Sprintf("Worker: %d ms", elapsed.Milliseconds()))
			}(item)
		}
	}()
}

func (d *DurableDaemon) runMaintenanceServer() error {
	d.maintenanceSrv = &http.Server{Addr: fmt.Sprintf(":%d", d.maintenancePort)}

	setMaintenanceServerHandlers(d.maintenancePath, d.pipelines, d.logger)

	durableQueueHandler := d.maintenancePath + "/queue"
	http.HandleFunc(durableQueueHandler, func(resp http.ResponseWriter, req *http.Request) {
		respMap := make(map[string]interface{})
		var unprocessed []map[string]interface{}
		var inflight  []map[string]interface{}
		unprocessedItems, err := d.dQueue.GetUnprocessed()
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}
		inflightItems, err := d.dQueue.GetInflight()
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}

		for _, item := range inflightItems {
			itemMap, err := common.JsonToMap(item.Data)
			if err != nil {
				resp.WriteHeader(500)
				_, _ = resp.Write([]byte(err.Error()))
				d.logger.Error(err.Error())
				return
			}
			inflight = append(inflight,
				map[string]interface{}{
				"data": itemMap,
				"queueTime": item.QueueTime,
				"inflightTime": item.InflightTime,
				"idx": item.Idx,
				})
		}

		for _, item := range unprocessedItems {
			itemMap, err := common.JsonToMap(item.Data)
			if err != nil {
				resp.WriteHeader(500)
				_, _ = resp.Write([]byte(err.Error()))
				d.logger.Error(err.Error())
				return
			}
			unprocessed = append(unprocessed,
				map[string]interface{}{
					"data": itemMap,
					"queueTime": item.QueueTime,
					"idx": item.Idx,
				})
		}

		respMap["inflight"] = inflight
		respMap["unprocessed"] = unprocessed

		respJson, err := common.MapToJson(respMap)
		if err != nil {
			resp.WriteHeader(500)
			_, _ = resp.Write([]byte(err.Error()))
			d.logger.Error(err.Error())
			return
		}
		resp.Header().Set("context-type", "application/json")
		resp.WriteHeader(200)
		_, _ = resp.Write(respJson)
	})

	return d.maintenanceSrv.ListenAndServe()
}

func (d *DurableDaemon) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", d.daemonPort))
	if err != nil {
		return err
	}

	defer l.Close()

	l = netutil.LimitListener(l, d.maxConnections)

	d.srv = &http.Server{}

	go func() {
		_ = d.runMaintenanceServer()
	}()

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
	if err := d.maintenanceSrv.Shutdown(context.Background()); err != nil {
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
