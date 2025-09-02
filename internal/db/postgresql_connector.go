package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresqlConnector struct{}

func (p *PostgresqlConnector) Connect(user string, password string, host string, port string, dbname string) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to %s database at %s:%s\n", dbname, host, port)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		fmt.Printf("Failed to ping %s database at %s:%s\n", dbname, host, port)
		return nil, err
	}

	return pool, nil
}
