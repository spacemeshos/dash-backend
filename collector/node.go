package collector

import (
    "context"
    "io"
    "time"

//     pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
//    "google.golang.org/grpc"
//    "google.golang.org/grpc/grpclog"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"

//    "github.com/spacemeshos/dash-backend/types"
)

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
        c.notify <- -streamType_node_SyncStatus
        log.Info("Stop node sync status pump")
    }()

    c.notify <- +streamType_node_SyncStatus

    stream, err := c.nodeClient.SyncStatusStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get sync status stream: %s", err)
        return err
    }

    c.syncStart()

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

//        switch res.GetStatus() {
//        case pb.NodeSyncStatus_NOT_SYNCED:
//            c.syncStart()
//        }
    }

    return nil
}

func (c *Collector) errorPump() error {
    var req empty.Empty

    log.Info("Start node error pump")
    defer func() {
        c.notify <- -streamType_node_Error
        log.Info("Stop node error pump")
    }()

    c.notify <- +streamType_node_Error

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
