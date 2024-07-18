package psql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Req struct {
	ctx  context.Context
	pool *pgxpool.Pool
	queryTimeOut int
}

func NewReq(ctx context.Context, pool *pgxpool.Pool, queryTimeOut int) *Req {
	return &Req{
		ctx:  ctx,
		pool: pool,
		queryTimeOut: queryTimeOut,
	}
}
