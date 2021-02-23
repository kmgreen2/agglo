package main

import (
	"context"
	gocrypto "crypto"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/observability"
	"github.com/kmgreen2/agglo/pkg/util"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

type TickerServer struct {
	grpcPort                int
	httpPort                int
	exporter                observability.Exporter
	kvStoreConnectionString string
	kvStore                 kvs.KVStore
	ticker                  entwine.TickerStore
	authenticators          map[string]crypto.Authenticator
	signer 					crypto.Signer
	api.UnimplementedTickerServer
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() (*TickerServer, error) {
	var err error
	server := &TickerServer{}
	grpcPortPtr := flag.Int("grpcPort", 9081, "grpc listening port (default 9081)")
	httpPortPtr := flag.Int("httpPort", 9080, "http listening port (default 9080)")
	exporterPtr := flag.String("exporter", "stdout", "OpenTelemetry exporter type")
	kvConnectionStringPtr := flag.String("kvConnectionString", "mem:local",
		"KVStore connection string (default:  mem:local)")
	authenticatorPathPtr := flag.String("authenticatorPath", "/etc/entwine/authenticators",
		"path to directory containing public keys for authenticators")
	privateKeyPathPtr := flag.String("privateKeyPath", "/etc/entwine/tickerKey.pem",
		"path to PEM filer for ticker")

	flag.Parse()

	server.grpcPort = *grpcPortPtr
	server.httpPort = *httpPortPtr

	kvStore, err := kvs.NewKVStoreFromConnectionString(*kvConnectionStringPtr)
	if err != nil {
		 return nil, err
	}

	server.ticker = entwine.NewKVTickerStore(kvStore, util.SHA256)

	if strings.Compare(*exporterPtr, "stdout") == 0 {
		server.exporter, err = observability.NewStdoutExporter()
		if err != nil {
			panic(err)
		}
	} else if strings.Compare(*exporterPtr, "zipkin") == 0 {
		server.exporter = observability.NewZipkinExporter("http://localhost:9411/api/v2/spans", "binge")
	} else {
		usage(fmt.Sprintf("invalid exporter: %s", *exporterPtr), 1)
	}

	files, err := ioutil.ReadDir(*authenticatorPathPtr)
	if err != nil {
		return nil, err
	}

	server.authenticators = make(map[string]crypto.Authenticator)

	for _, file := range files {
		fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", *authenticatorPathPtr, file.Name()))
		if err != nil {
			return nil, err
		}
		publicKey, err := crypto.ParseRSAPublicKeyFromPEM(string(fileBytes))
		if err != nil {
			return nil, err
		}
		server.authenticators[file.Name()] = crypto.NewRSAAuthenticator(publicKey, gocrypto.SHA256)
	}

	fileBytes, err := ioutil.ReadFile(*privateKeyPathPtr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(string(fileBytes))
	if err != nil {
		return nil, err
	}

	server.signer = crypto.NewRSASigner(privateKey, gocrypto.SHA256)

	return server, nil
}

func startGrpcServer(server *TickerServer) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", server.grpcPort))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	api.RegisterTickerServer(grpcServer, server)
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server, err := parseArgs()
	if err != nil {
		panic(err)
	}

	err = startGrpcServer(server)
	if err != nil {
		panic(err)
	}
	// Register gRPC server endpoint
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = api.RegisterTickerHandlerFromEndpoint(ctx, mux, fmt.Sprintf(":%d", server.grpcPort), opts)
	if err != nil {
		panic(err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(fmt.Sprintf(":%d", server.httpPort), mux)
	if err != nil {
		panic(err)
	}
}

func (server *TickerServer) Anchor(ctx context.Context, request *api.AnchorRequest) (*api.AnchorResponse, error) {
	var authenticator crypto.Authenticator
	var proof []*entwine.StreamImmutableMessage
	if _, ok := server.authenticators[request.SubStreamID]; !ok {
		return nil, util.NewInvalidError(fmt.Sprintf("could not find authenitcator for subStreamID=%s",
			request.SubStreamID))
	} else {
		authenticator = server.authenticators[request.SubStreamID]
	}

	for _, pbMessage := range request.Proof {
		message, err := entwine.NewStreamImmutableMessageFromPb(pbMessage)
		if err != nil {
			return nil, err
		}
		proof = append(proof, message)
	}

	tickerMessage, err := server.ticker.Anchor(proof, entwine.SubStreamID(request.SubStreamID), authenticator)
	if err != nil {
		return nil, err
	}
	return &api.AnchorResponse{
		TickerMessage: entwine.NewPbFromTickerImmutableMessage(tickerMessage),
	}, nil
}

func (server *TickerServer) Tick(ctx context.Context, request *api.TickRequest) (*api.TickResponse, error) {
	err := server.ticker.Append(server.signer)
	if err != nil {
		return nil, err
	}
	return &api.TickResponse{}, nil
}

func (server *TickerServer) GetProofStartUuid(ctx context.Context,
	request *api.GetProofStartUuidRequest) (*api.GetProofStartUuidResponse, error) {

	uuid, err := server.ticker.GetProofStartUuid(entwine.SubStreamID(request.SubStreamID))
	if err != nil {
		return nil, err
	}
	return &api.GetProofStartUuidResponse{
		Uuid: uuid.String(),
	}, nil
}

func (server *TickerServer) CreateGenesisProof(ctx context.Context,
	request *api.CreateGenesisProofRequest) (*api.CreateGenesisProofResponse, error) {

	proof, err := server.ticker.CreateGenesisProof(entwine.SubStreamID(request.SubStreamID))
	if err != nil {
		return nil, err
	}
	return &api.CreateGenesisProofResponse{
		Uuid: proof.TickerUuid().String(),
	}, nil
}

