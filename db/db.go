package db

import (
	"context"
"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoClient(connectionUri string) (*mongo.Client, error) {
	ctx, cancel := DBReqContext(20)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionUri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func DBReqContext(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration*time.Second)
}