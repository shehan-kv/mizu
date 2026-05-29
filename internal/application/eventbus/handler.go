package eventbus

import (
	"context"
	"mizu/internal/domain/common"
)

type EventHandler func(ctx context.Context, event common.Event) error
