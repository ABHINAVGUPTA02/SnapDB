package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresqlConnector struct{}

func (p *PostgresqlConnector) Connect(user string, password string, host string, port string, dbname string) (db *pgxpool.Pool, err error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to %s database at %s:%s\n", dbname, host, port)
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		fmt.Printf("Failed to ping %s database at %s:%s\n", dbname, host, port)
		return nil, err
	}

	return pool, nil
}
