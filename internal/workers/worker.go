package workers

import (
	"context"

	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type Worker interface {
	Handle(ctx context.Context, event messaging.Event) error
}
