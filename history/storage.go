package history

import (
    "context"
    "time"
//    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type storage struct {
    client *mongo.Client
    db *mongo.Database
}

func (s *storage) open() error {
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

    if err != nil {
        return err
    }

    err = client.Ping(ctx, nil)

    if err != nil {
        return err
    }

    s.client = client
    s.db = client.Database("dashboard")

    return nil
}

func (s *storage) close() {
    if s.client != nil {
        ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
        s.db = nil
        s.client.Disconnect(ctx)
    }
}