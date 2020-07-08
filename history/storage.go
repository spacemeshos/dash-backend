package history

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"
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

func (s *storage) getEpoch(number uint64) (*Epoch, error) {
    query := bson.D{{"number", number}}
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("epoches").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    epoch := &Epoch{
        number: number,
        confirmed: doc.Lookup("confirmed").Boolean(),
    }
    stats := doc.Lookup("stats").Document()
    epoch.stats.capacity = uint64(stats.Lookup("capacity").Int64())
    epoch.stats.decentral = uint64(stats.Lookup("decentral").Int64())
    epoch.stats.smeshers = uint64(stats.Lookup("smeshers").Int64())
    epoch.stats.transactions = uint64(stats.Lookup("transactions").Int64())
    epoch.stats.accounts = uint64(stats.Lookup("accounts").Int64())
    epoch.stats.circulation = uint64(stats.Lookup("circulation").Int64())
    epoch.stats.rewards = uint64(stats.Lookup("rewards").Int64())
    epoch.stats.security = uint64(stats.Lookup("security").Int64())
    return epoch, nil
}

func (s *storage) putEpoch(epoch *Epoch) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("epoches").InsertOne(ctx, bson.D{
        {"number", epoch.number},
        {"confirmed", epoch.confirmed},
        {"stats", bson.D{
            {"capacity", epoch.stats.capacity},
            {"decentral", epoch.stats.decentral},
            {"smeshers", epoch.stats.smeshers},
            {"transactions", epoch.stats.transactions},
            {"accounts", epoch.stats.accounts},
            {"circulation", epoch.stats.circulation},
            {"rewards", epoch.stats.rewards},
            {"security", epoch.stats.security},
        }},
    })
    if err != nil {
        return err
    }
    id := res.InsertedID
    log.Info("putEpoch _id: %v", id)
    return nil
}

func (s *storage) close() {
    if s.client != nil {
        ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
        s.db = nil
        s.client.Disconnect(ctx)
    }
}