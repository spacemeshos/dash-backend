package collector

import (
    "context"
    "io"

    pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    "google.golang.org/grpc"
//    "google.golang.org/grpc/grpclog"
    "github.com/golang/protobuf/ptypes/empty"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/types"
)

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
        l, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive layer: %s", err)
            return err
        }

        log.Info("Mesh stream: %s", l.GetNumber())
        blocks := l.GetBlocks()
        layer := &types.Layer{
            Index: sm.LayerID(l.GetNumber()),
            Status: l.GetStatus(),
            Hash: l.GetHash(),
            Blocks: make([]*types.Block, len(blocks)),
            RootStateHash: l.GetRootStateHash(),
        }

        for i, b := range blocks {
            var id sm.BlockID
            copy(id[:], b.GetId())
            txs := b.GetTransactions()
            block := &types.Block{
                Id: id,
                Transactions: make([]*types.Transaction, len(txs)),
            }
            for j, t := range txs {
                var id sm.TransactionID
                var atx sm.TransactionID
                copy(id[:], t.GetId().GetId())
                copy(atx[:], t.GetPrevAtx().GetId())
                tx := &types.Transaction{
                    TxType: t.GetType(),
                    Id: id,
                    Sender: sm.BytesToAddress(t.GetSender().GetAddress()),
                    Fee: types.TransactionFee{
                        Gas_consumed: t.GetFee().GetGasConsumed(),
                        Gas_price: t.GetFee().GetGasPrice(),
                    },
                    Timestamp: t.GetTimestamp(),
                    Receiver: sm.BytesToAddress(t.GetReceiver().GetAddress()),
                    Amount: types.Amount(t.GetAmount().GetValue()),
                    Counter: t.GetCounter(),
                    Data: t.GetData(),
                    Smesher_id: sm.BytesToAddress(t.GetSmesherId().GetId()),
                    Prev_atx: atx,
                    Result: pb.TransactionReceipt_UNKNOWN,
                }
                block.Transactions[j] = tx
            }
            layer.Blocks[i] = block
        }

        c.history.AddLayer(layer)
    }

    return nil
}
