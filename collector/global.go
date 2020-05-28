package collector

import (
    "context"
    "io"

//     pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/types"
)

func (c *Collector) accountsPump() error {
    var req empty.Empty

    log.Info("Start global state account pump")
    defer func() {
        c.notify <- -streamType_global_Account
        log.Info("Stop global state account pump")
    }()

    c.notify <- +streamType_global_Account

    stream, err := c.globalClient.AccountStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state account stream: %s", err)
        return err
    }

    for {
        account, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive Account: %s", err)
            return err
        }

        log.Info("Account: %s", account.GetAddress())
        c.history.AddAccount(
            &types.Account{
                Address: sm.BytesToAddress(account.GetAddress().GetAddress()),
                Counter: account.GetCounter(),
                Balance: types.Amount(account.GetBalance().GetValue()),
            },
        )
    }

    return nil
}

func (c *Collector) rewardsPump() error {
    var req empty.Empty

    log.Info("Start global state reward pump")
    defer func() {
        c.notify <- -streamType_global_Reward
        log.Info("Stop global state reward pump")
    }()

    c.notify <- +streamType_global_Reward

    stream, err := c.globalClient.RewardStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state reward stream: %s", err)
        return err
    }

    for {
        reward, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive Reward: %s", err)
            return err
        }

        var smesherId types.SmesherID
        copy(smesherId[:], reward.GetSmesher().GetId())

        log.Info("Reward: %s", reward.GetTotal())
        c.history.AddReward(
            &types.Reward{
                Layer: sm.LayerID(reward.GetLayer()),
                Total: types.Amount(reward.GetTotal().GetValue()),
                Layer_reward: types.Amount(reward.GetLayerReward().GetValue()),
                Layer_computed: sm.LayerID(reward.GetLayerComputed()),
                Coinbase: sm.BytesToAddress(reward.GetCoinbase().GetAddress()),
                Smesher: smesherId,
            },
        )
    }

    return nil
}

func (c *Collector) transactionsReceiptPump() error {
    var req empty.Empty

    log.Info("Start global transactions receipt pump")
    defer func() {
        c.notify <- -streamType_global_TransactionReceipt
        log.Info("Stop global transactions receipt pump")
    }()

    c.notify <- +streamType_global_TransactionReceipt

    stream, err := c.globalClient.TransactionReceiptStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global transactions receipt stream: %s", err)
        return err
    }

    for {
        txReceipt, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive TransactionReceipt: %s", err)
            return err
        }

        log.Info("TransactionReceipt: %s", txReceipt.GetId())
        var id sm.TransactionID
        copy(id[:], txReceipt.GetId().GetId())

        c.history.AddTransactionReceipt(
            &types.TransactionReceipt{
                Id: id,
                Result: txReceipt.GetResult(),
                Gas_used: txReceipt.GetGasUsed(),
                Fee: types.Amount(txReceipt.GetFee().GetValue()),
                Layer_number: sm.LayerID(txReceipt.GetLayerNumber()),
            },
        )
    }

    return nil
}

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
        log.Error("cannot get global state transactions state: %s", err)
        return err
    }

    for {
        txState, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive TransactionState: %s", err)
            return err
        }

        log.Info("TransactionState: %s, %s", txState.GetId(), txState.GetState())
        var id sm.TransactionID
        copy(id[:], txState.GetId().GetId())

        c.history.AddTransactionState(&id, txState.GetState());
    }

    return nil
}

