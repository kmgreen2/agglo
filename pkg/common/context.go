package common

import "context"

func ExtractPubSubContext(payload []byte) context.Context {
	return context.Background()
}

func InjectPubSubContext(ctx context.Context, payload []byte) context.Context {
	return context.Background()
}
