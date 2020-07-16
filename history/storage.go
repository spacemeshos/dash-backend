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
    current := stats.Lookup("current").Document()
    epoch.stats.current.capacity = uint64(current.Lookup("capacity").Int64())
    epoch.stats.current.decentral = uint64(current.Lookup("decentral").Int64())
    epoch.stats.current.smeshers = uint64(current.Lookup("smeshers").Int64())
    epoch.stats.current.transactions = uint64(current.Lookup("transactions").Int64())
    epoch.stats.current.accounts = uint64(current.Lookup("accounts").Int64())
    epoch.stats.current.circulation = uint64(current.Lookup("circulation").Int64())
    epoch.stats.current.rewards = uint64(current.Lookup("rewards").Int64())
    epoch.stats.current.security = uint64(current.Lookup("security").Int64())
    cumulative := stats.Lookup("cumulative").Document()
    epoch.stats.cumulative.capacity = uint64(cumulative.Lookup("capacity").Int64())
    epoch.stats.cumulative.decentral = uint64(cumulative.Lookup("decentral").Int64())
    epoch.stats.cumulative.smeshers = uint64(cumulative.Lookup("smeshers").Int64())
    epoch.stats.cumulative.transactions = uint64(cumulative.Lookup("transactions").Int64())
    epoch.stats.cumulative.accounts = uint64(cumulative.Lookup("accounts").Int64())
    epoch.stats.cumulative.circulation = uint64(cumulative.Lookup("circulation").Int64())
    epoch.stats.cumulative.rewards = uint64(cumulative.Lookup("rewards").Int64())
    epoch.stats.cumulative.security = uint64(cumulative.Lookup("security").Int64())
    return epoch, nil
}

func (s *storage) putEpoch(epoch *Epoch) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("epoches").InsertOne(ctx, bson.D{
        {"number", epoch.number},
        {"confirmed", epoch.confirmed},
        {"stats", bson.D{
            {"current",  bson.D{
                {"capacity", epoch.stats.current.capacity},
                {"decentral", epoch.stats.current.decentral},
                {"smeshers", epoch.stats.current.smeshers},
                {"transactions", epoch.stats.current.transactions},
                {"accounts", epoch.stats.current.accounts},
                {"circulation", epoch.stats.current.circulation},
                {"rewards", epoch.stats.current.rewards},
                {"security", epoch.stats.current.security},
            }},
            {"cumulative",  bson.D{
                {"capacity", epoch.stats.cumulative.capacity},
                {"decentral", epoch.stats.cumulative.decentral},
                {"smeshers", epoch.stats.cumulative.smeshers},
                {"transactions", epoch.stats.cumulative.transactions},
                {"accounts", epoch.stats.cumulative.accounts},
                {"circulation", epoch.stats.cumulative.circulation},
                {"rewards", epoch.stats.cumulative.rewards},
                {"security", epoch.stats.cumulative.security},
            }},
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