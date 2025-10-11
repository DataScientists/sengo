package repositoryutil

import (
	"context"
	"sheng-go-backend/ent"
)

func WithTransactionalMutation(ctx context.Context) *ent.Client {
	return ent.FromContext(ctx)
}
