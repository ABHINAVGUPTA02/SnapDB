package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnector struct{}

func (m *MongoConnector) Connect(user string, password string, host string, port string, dbname string) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", user, password, host, port, dbname)
	fmt.Printf("Connecting to MongoDB database %s at %s:%s\n", dbname, host, port)
	clientOptions := options.Client().ApplyURI(uri)

	// connecting with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("Failed to connect to %s database %s at %s:%s\n", dbname, dbname, host, port)
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Printf("Failed to ping to %s database %s at %s:%s\n", dbname, dbname, host, port)
		return nil, err
	}

	return client, nil
}
