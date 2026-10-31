package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDeps groups the MongoDB client and the selected database.
type MongoDeps struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongo connects to MongoDB and verifies the connection with ping.
func ConnectMongo(ctx context.Context, uri, dbName string) (*MongoDeps, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(dialCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, 3*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return &MongoDeps{Client: client, Database: client.Database(dbName)}, nil
}
