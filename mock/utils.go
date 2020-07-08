package mock

import (
    "math/rand"
//    "time"

    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    "github.com/spacemeshos/ed25519"
//    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func randUint(min int, max int) int {
    return min + rand.Intn(max-min)
}

func newActivationID(id *ActivationID) {
    for i := 0; i < len(id); i++ {
        id[i] = (byte)(rand.Intn(255))
    }
}

func newSmesherID(id *SmesherID) {
    for i := 0; i < len(id); i++ {
        id[i] = (byte)(rand.Intn(255))
    }
}

func newTransactionID(id *sm.TransactionID) {
    for i := 0; i < len(id); i++ {
        id[i] = (byte)(rand.Intn(255))
    }
}

func newBlockID(id *sm.BlockID) {
    for i := 0; i < len(id); i++ {
        id[i] = (byte)(rand.Intn(255))
    }
}

func getRate(rates []Rate) int {
    var total int
    for _, r  := range rates {
        total += r.Rate
    }
    t := rand.Intn(total)
    total = 0
    for _, r  := range rates {
        total += r.Rate
        if t < total {
            return r.Value
        }
    }
    return rates[len(rates) - 1].Value
}

