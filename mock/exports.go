package mock

import (
//    "math/rand"
//    "time"

    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    "github.com/spacemeshos/ed25519"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (id *ActivationID) Bytes() []byte { return id[:] }
func (id *SmesherID) Bytes() []byte { return id[:] }
func getBlockIdBytes(id *sm.BlockID) []byte { return id[:] }

func (account *Account) Export() *pb.Account {
    return &pb.Account{
        Address: &pb.AccountId{Address: account.Address.Bytes()},
        Counter: account.Counter,
        Balance: &pb.Amount{Value: uint64(account.Balance)},
    }
}

func (reward *Reward) Export() *pb.Reward {
    return &pb.Reward{
        Layer: uint64(reward.Layer),
        Total: &pb.Amount{Value: uint64(reward.Total)},
        LayerReward: &pb.Amount{Value: uint64(reward.LayerReward)},
        LayerComputed: uint64(reward.LayerComputed),
        Coinbase: &pb.AccountId{Address: reward.Coinbase.Bytes()},
        Smesher: &pb.SmesherId{Id: reward.Smesher.Bytes()},
    }
}

func (gasOffered *GasOffered) Export() *pb.GasOffered {
    return &pb.GasOffered{
        GasProvided: gasOffered.GasProvided,
        GasPrice: gasOffered.GasPrice,
    }
}

func (signature *Signature) Export() *pb.Signature {
    return &pb.Signature{
        Scheme: signature.Scheme,
        Signature: signature.Signature,
        PublicKey: signature.PublicKey,
    }
}

func (tx *Transaction) Export() *pb.Transaction {
    pbTx := &pb.Transaction{
        Id: &pb.TransactionId{Id: tx.Id.Bytes()},
        Sender: &pb.AccountId{Address: tx.Sender.Bytes()},
        GasOffered: tx.GasOffered.Export(),
        Amount: &pb.Amount{Value: uint64(tx.Amount)},
        Counter: tx.Counter,
        Signature:  tx.Signature.Export(),
    }
    if tx.IsSimple {
        pbTx.Data = &pb.Transaction_CoinTransfer{
            CoinTransfer: &pb.CoinTransferTransaction{
                Receiver: &pb.AccountId{Address: tx.Receiver.Bytes()},
            },
        }
    } else {
        pbTx.Data = &pb.Transaction_SmartContract{
            SmartContract: &pb.SmartContractTransaction{
                Type: tx.Type,
                Data: tx.Data,
                Address: &pb.AccountId{Address: tx.Address.Bytes()},
            },
        }
    }
    return pbTx
}

func (layer *Layer) Export() *pb.Layer {
    blocks := make([]*pb.Block, 0, len(layer.Blocks))
    atxs := make([]*pb.Activation, 0, len(layer.Activations))
    for _, block := range layer.Blocks {
        blocks = append(blocks, block.Export())
    }
    for _, atx := range layer.Activations {
        atxs = append(atxs, atx.Export())
    }
    return &pb.Layer{
        Number: uint64(layer.Number),
        Status: layer.Status,
        Hash: layer.Hash,
        Blocks: blocks,
        Activations: atxs,
        RootStateHash: layer.RootStateHash,
    }
}

func (block *Block) Export() *pb.Block {
    txs := make([]*pb.Transaction, 0, len(block.Transactions))
    for _, tx := range block.Transactions {
        txs = append(txs, tx.Export())
    }
    return &pb.Block{
        Id: getBlockIdBytes(&block.Id),
        Transactions: txs,
    }
}

func (atx *Activation) Export() *pb.Activation {
    return &pb.Activation{
        Id: &pb.ActivationId{Id: atx.Id.Bytes()},
        Layer: uint64(atx.Layer),
        SmesherId: &pb.SmesherId{Id: atx.SmesherId.Bytes()},
        Coinbase: &pb.AccountId{Address: atx.Coinbase.Bytes()},
        PrevAtx: &pb.ActivationId{Id: atx.PrevAtx.Bytes()},
        CommitmentSize: atx.CommitmentSize,
    }
}

