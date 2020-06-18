package collector

import (
    "context"
    "io"
    "time"

//    pb "github.com/spacemeshos/dash-backend/spacemesh/v1"
//    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    "google.golang.org/grpc"
//    "google.golang.org/grpc/grpclog"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/types"
)

func (c *Collector) getNetworkInfo() error {
    var req empty.Empty

    // set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    genesisTime, err := c.meshClient.GenesisTime(ctx, &req)
    if err != nil {
        log.Error("cannot get GenesisTime: %s", err)
        return err
    }

    netId, err := c.meshClient.NetId(ctx, &req)
    if err != nil {
        log.Error("cannot get NetId: %s", err)
        return err
    }

    epochNumLayers, err := c.meshClient.EpochNumLayers(ctx, &req)
    if err != nil {
        log.Error("cannot get EpochNumLayers: %s", err)
        return err
    }

    maxTransactionsPerSecond, err := c.meshClient.MaxTransactionsPerSecond(ctx, &req)
    if err != nil {
        log.Error("cannot get MaxTransactionsPerSecond: %s", err)
        return err
    }

    c.history.SetNetworkInfo(
        netId.GetNetid().GetValue(),
        genesisTime.GetUnixtime().GetValue(),
        epochNumLayers.GetNumlayers().GetValue(),
        maxTransactionsPerSecond.GetMaxtxpersecond().GetValue(),
        15,
    )

    return nil
}

func (c *Collector) layersPump() error {
    var req empty.Empty

    log.Info("Start mesh layer pump")
    defer func() {
        c.notify <- -streamType_mesh_Layer
        log.Info("Stop mesh layer pump")
    }()

    c.notify <- +streamType_mesh_Layer

    stream, err := c.meshClient.LayerStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get layer stream: %s", err)
        return err
    }

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive layer: %s", err)
            return err
        }
        layer := response.GetLayer()
        log.Info("Mesh stream: %s", layer.GetNumber())
        c.history.AddLayer(types.NewLayer(layer))
    }

    return nil
}
