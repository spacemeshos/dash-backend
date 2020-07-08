package collector

import (
    "context"
    "io"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
//    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/types"
)

func (c *Collector) globalStatePump() error {
    req := pb.GlobalStateStreamRequest{GlobalStateDataItemFlags: uint32(pb.GlobalStateDataItemFlag_GLOBAL_STATE_DATA_ITEM_FLAG_REWARD) | uint32(pb.GlobalStateDataItemFlag_GLOBAL_STATE_DATA_ITEM_FLAG_ACCOUNT)}

    log.Info("Start global state pump")
    defer func() {
        c.notify <- -streamType_globalState
        log.Info("Stop global state pump")
    }()

    c.notify <- +streamType_globalState

    stream, err := c.globalClient.GlobalStateStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state account stream: %v", err)
        return err
    }

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive Global state data: %v", err)
            return err
        }
        for _, item := range response.GetDataItem() {
            if account := item.GetAccount(); account != nil {
                c.history.AddAccount(types.NewAccount(account))
            } else if reward := item.GetReward(); reward != nil {
                c.history.AddReward(types.NewReward(reward))
            } else if receipt := item.GetReceipt(); receipt != nil {
                c.history.AddTransactionReceipt(types.NewTransactionReceipt(receipt))
            }
        }
    }

    return nil
}
/*
func (c *Collector) transactionsStatePump() error {
    var req empty.Empty

    log.Info("Start global state transactions state pump")
    defer func() {
        c.notify <- -streamType_global_TransactionState
        log.Info("Stop global state transactions state pump")
    }()

    c.notify <- +streamType_global_TransactionState

    stream, err := c.globalClient.TransactionStateStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state transactions state: %v", err)
        return err
    }

    for {
        txState, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive TransactionState: %v", err)
            return err
        }

        log.Info("TransactionState: %v, %v", txState.GetId(), txState.GetState())
        var id sm.TransactionID
        copy(id[:], txState.GetId().GetId())

        c.history.AddTransactionState(&id, txState.GetState());
    }

    return nil
}
*/
