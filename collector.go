package main

import (
    "context"
    "io"
    "time"

     pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
    "google.golang.org/grpc"
//    "google.golang.org/grpc/grpclog"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"
)

type Collector struct {
    apiUrl	string
    history	*History
    nodeClient	pb.NodeServiceClient

    streams [2]bool
    activeStreams int
    online bool

    // Stream status changed.
    notify chan int
}

func (c *Collector) syncStart() error {
    var req empty.Empty

    // set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := c.nodeClient.SyncStart(ctx, &req)
    if err != nil {
        log.Error("cannot create laptop: ", err)
        return err
    }

    log.Info("Started node sync")
    return nil
}

func (c *Collector) syncStatusPump() error {
    var req empty.Empty

    log.Info("Start node sync status pump")
    defer func() {
        c.notify <- -1
        log.Info("Stop node sync status pump")
    }()

    c.notify <- 1

    stream, err := c.nodeClient.SyncStatusStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get sync status stream: %s", err)
        return err
    }

    for {
        res, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive sync status: %s", err)
            return err
        }

        log.Info("Node sync status: %s", res.GetStatus())

        switch res.GetStatus() {
        case pb.NodeSyncStatus_NOT_SYNCED:
            c.syncStart()
        }
    }

    return nil
}

func (c *Collector) errorPump() error {
    var req empty.Empty

    log.Info("Start node error pump")
    defer func() {
        c.notify <- -2
        log.Info("Stop node error pump")
    }()

    c.notify <- 2

    stream, err := c.nodeClient.ErrorStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get error stream: %s", err)
        return err
    }

    for {
        res, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive error: %s", err)
            return err
        }

        log.Info("Node error: %s", res.GetMessage())
    }

    return nil
}

func NewCollector(nodeAddress string, history *History) *Collector {
    return &Collector{
        apiUrl:  nodeAddress,
        history: history,
        notify:  make(chan int),
    }
}

func (c *Collector) Run() {
    for {
        log.Info("dial node %s", c.apiUrl)

        conn, err := grpc.Dial(c.apiUrl, grpc.WithInsecure())
        if err != nil {
            log.Error("cannot dial node: %s", err)
            time.Sleep(5 * time.Second)
            continue
        }

        c.nodeClient = pb.NewNodeServiceClient(conn)

        go c.syncStatusPump()
        go c.errorPump()

        for {
            state := <-c.notify
            switch {
            case state > 0:
                c.streams[state - 1] = true
                c.activeStreams++
            case state < 0:
                c.streams[(-state) - 1] = false
                c.activeStreams--
            }
            if c.activeStreams == 2 {
                c.online = true
            }
            if c.online && c.activeStreams < 2 {
                break
            }
        }

        time.Sleep(5 * time.Second)
    }
}
