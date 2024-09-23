package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConnection = fmt.Errorf("error dial to box")
)

type Connection struct {
	pool *pgxpool.Pool
	opts *ConnectionOptions
	done chan struct{}
}

func GetConnection(ctx context.Context, postgresOpts *ConnectionOptions) (*Connection, error) {
	pool, err := pgxpool.NewWithConfig(ctx, &postgresOpts.poolCfg) // ToDO target_session_attrs for replica/master/standby mode
	if err != nil {
		return nil, fmt.Errorf("%w %s with connect timeout '%d': %s", ErrConnection, postgresOpts.poolCfg.ConnConfig.Host, postgresOpts.poolCfg.ConnConfig.ConnectTimeout, err)
	}

	return &Connection{pool: pool, opts: postgresOpts, done: make(chan struct{})}, nil
}

func (c *Connection) Call(ctx context.Context, sql string, args []any) (pgx.Rows, error) {
	if c == nil || c.pool == nil {
		return nil, fmt.Errorf("attempt call from empty connection")
	}

	return c.pool.Query(ctx, sql, args...)
}

func (c *Connection) InstanceMode() any {
	return c.opts.InstanceMode()
}

func (c *Connection) Close() {
	if c == nil || c.pool == nil {
		return
	}

	go func() {
		c.pool.Close()
		c.pool = nil
		c.done <- struct{}{}
	}()
}

func (c *Connection) Done() <-chan struct{} {
	return c.done
}

func (c *Connection) Info() string {
	return fmt.Sprintf("Server: %s, timeout; %d, poolSize: %d", c.opts.poolCfg.ConnConfig.Host, c.opts.poolCfg.ConnConfig.ConnectTimeout, c.opts.poolCfg.MaxConns)
}
