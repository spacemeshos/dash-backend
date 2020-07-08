package types

import (
//    "encoding/json"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func NewAccount(account *pb.Account) *Account {
    return &Account{
        Address: sm.BytesToAddress(account.GetAddress().GetAddress()),
        Counter: account.GetCounter(),
        Balance: Amount(account.GetBalance().GetValue()),
    }
}

func NewReward(reward *pb.Reward) *Reward {
    var smesherId SmesherID
    copy(smesherId[:], reward.GetSmesher().GetId())
    return &Reward{
        Layer: sm.LayerID(reward.GetLayer()),
        Total: Amount(reward.GetTotal().GetValue()),
        Layer_reward: Amount(reward.GetLayerReward().GetValue()),
        Layer_computed: sm.LayerID(reward.GetLayerComputed()),
        Coinbase: sm.BytesToAddress(reward.GetCoinbase().GetAddress()),
        Smesher: smesherId,
    }
}

func NewTransactionReceipt(txReceipt *pb.TransactionReceipt) *TransactionReceipt {
    var id sm.TransactionID
    copy(id[:], txReceipt.GetId().GetId())
    return &TransactionReceipt{
        Id: id,
        Result: txReceipt.GetResult(),
        Gas_used: txReceipt.GetGasUsed(),
        Fee: Amount(txReceipt.GetFee().GetValue()),
        Layer_number: sm.LayerID(txReceipt.GetLayerNumber()),
        Index: txReceipt.GetIndex(),
        App_address: sm.BytesToAddress(txReceipt.GetAppAddress().GetAddress()),
    }
}

func NewSignature(sig *pb.Signature) Signature {
    return Signature{
        Scheme: sig.GetScheme(),
        Signature: sig.GetSignature(),
        Public_key: sig.GetPublicKey(),
    }
}

func NewGasOffered(gas *pb.GasOffered) GasOffered {
    return GasOffered{
        Gas_provided: gas.GetGasProvided(),
        Gas_price: gas.GetGasPrice(),
    }
}

func NewActivation(atx *pb.Activation) *Activation {
    var id ActivationID
    copy(id[:], atx.GetId().GetId())
    var smesherId SmesherID
    copy(smesherId[:], atx.GetSmesherId().GetId())
    var prevAtxId ActivationID
    copy(prevAtxId[:], atx.GetPrevAtx().GetId())
    return &Activation{
        Id: id,
        Layer: sm.LayerID(atx.GetLayer()),
        Smesher_id: smesherId,
        Coinbase: sm.BytesToAddress(atx.GetCoinbase().GetAddress()),
        Prev_atx: prevAtxId,
        Commitment_size: atx.GetCommitmentSize(),
    }
}

func NewLayer(l *pb.Layer) *Layer {
    blocks := l.GetBlocks()
    atxs := l.GetActivations()
    layer := &Layer{
        Number: sm.LayerID(l.GetNumber()),
        Status: l.GetStatus(),
        Hash: l.GetHash(),
        Blocks: make([]*Block, len(blocks)),
        Activations: make([]*Activation, len(atxs)),
        RootStateHash: l.GetRootStateHash(),
        Transactions: make(map[sm.TransactionID]Transaction),
    }

    for i, b := range blocks {
        var id sm.BlockID
        copy(id[:], b.GetId())
        txs := b.GetTransactions()
        block := &Block{
            Id: id,
            Transactions: make([]Transaction, len(txs)),
        }
        for j, t := range txs {
            var id sm.TransactionID
            copy(id[:], t.GetId().GetId())
            if data := t.GetCoinTransfer(); data != nil {
                tx := &CoinTransferTransaction{
                    TransactionBase{
                        Id: id,
                        Sender: sm.BytesToAddress(t.GetSender().GetAddress()),
                        Gas_offered: NewGasOffered(t.GetGasOffered()),
                        Amount: Amount(t.GetAmount().GetValue()),
                        Counter: t.GetCounter(),
                        Signature: NewSignature(t.GetSignature()),
                    },
                    sm.BytesToAddress(data.GetReceiver().GetAddress()),
                }
                block.Transactions[j] = tx
                layer.Transactions[id] = tx
            } else if data := t.GetSmartContract(); data != nil {
                tx := &SmartContractTransaction{
                    TransactionBase{
                        Id: id,
                        Sender: sm.BytesToAddress(t.GetSender().GetAddress()),
                        Gas_offered: NewGasOffered(t.GetGasOffered()),
                        Amount: Amount(t.GetAmount().GetValue()),
                        Counter: t.GetCounter(),
                        Signature: NewSignature(t.GetSignature()),
                    },
                    data.GetType(),
                    data.GetData(),
                    sm.BytesToAddress(data.GetAddress().GetAddress()),
                }
                block.Transactions[j] = tx
                layer.Transactions[id] = tx
            }
        }
        layer.Blocks[i] = block
    }

    for i, a := range atxs {
        layer.Activations[i] = NewActivation(a)
    }

    return layer
}
