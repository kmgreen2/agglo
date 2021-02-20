package client

import (
	"context"
	gUuid "github.com/google/uuid"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"google.golang.org/grpc"
)

type TickerClientImpl struct {
	endpoint   string
	conn       *grpc.ClientConn
	underlying api.TickerClient
}

type TickerClient interface {
	Close() error
	Anchor(ctx context.Context, proof []*entwine.StreamImmutableMessage,
		subStreamID entwine.SubStreamID) (*entwine.TickerImmutableMessage, error)
	Tick(ctx context.Context) error
	GetProofStartUuid(ctx context.Context,
		subStreamID entwine.SubStreamID) (gUuid.UUID, error)
	CreateGenesisProof(ctx context.Context,
		subStreamID entwine.SubStreamID) (gUuid.UUID, error)
}

func NewTickerClient(endpoint string, opts... grpc.DialOption) (TickerClient, error) {
	var err error
	tickerClient := &TickerClientImpl{}
	tickerClient.conn, err = grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	tickerClient.underlying = api.NewTickerClient(tickerClient.conn)
	return tickerClient, nil
}

func (tickerClient *TickerClientImpl) Close() error {
	return tickerClient.conn.Close()
}

func (tickerClient *TickerClientImpl) Anchor(ctx context.Context, proof []*entwine.StreamImmutableMessage,
	subStreamID entwine.SubStreamID) (*entwine.TickerImmutableMessage, error) {

	var pbStreamImmutableMessages []*api.StreamImmutableMessage

	for _, message := range proof {
		pbStreamImmutableMessages = append(pbStreamImmutableMessages, entwine.NewPbFromStreamImmutableMessage(message))
	}

	response, err := tickerClient.underlying.Anchor(ctx, &api.AnchorRequest{
		Proof: pbStreamImmutableMessages,
		SubStreamID: string(subStreamID),
	})

	if err != nil {
		return nil, err
	}

	tickerMessage, err := entwine.NewTickerImmutableMessageFromPb(response.TickerMessage)
	if err != nil {
		return nil, err
	}

	return tickerMessage, nil
}

func (tickerClient *TickerClientImpl) Tick(ctx context.Context) error {
	_, err := tickerClient.underlying.Tick(ctx, &api.TickRequest{})
	if err != nil {
		return err
	}
	return nil
}

func (tickerClient *TickerClientImpl) GetProofStartUuid(ctx context.Context,
	subStreamID entwine.SubStreamID) (gUuid.UUID, error) {

	response, err := tickerClient.underlying.GetProofStartUuid(ctx, &api.GetProofStartUuidRequest{
		SubStreamID: string(subStreamID),
	})

	if err != nil {
		return gUuid.Nil, err
	}

	uuid, err := gUuid.Parse(response.Uuid)
	if err != nil {
		return gUuid.Nil, err
	}

	return uuid, nil
}

func (tickerClient *TickerClientImpl) CreateGenesisProof(ctx context.Context,
	subStreamID entwine.SubStreamID) (gUuid.UUID, error) {

	response, err := tickerClient.underlying.CreateGenesisProof(ctx, &api.CreateGenesisProofRequest{
		SubStreamID: string(subStreamID),
	})

	if err != nil {
		return gUuid.Nil, err
	}

	uuid, err := gUuid.Parse(response.Uuid)
	if err != nil {
		return gUuid.Nil, err
	}

	return uuid, nil
}


