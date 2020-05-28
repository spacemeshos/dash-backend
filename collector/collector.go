package collector

import (
    "time"

     pb "github.com/spacemeshos/dash-backend/spacemesh"
    "google.golang.org/grpc"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/history"
)

const (
    streamType_node_SyncStatus			int = 1
    streamType_node_Error			int = 2
    streamType_mesh_Layer			int = 3
    streamType_global_Account			int = 4
    streamType_global_Reward			int = 5
    streamType_global_TransactionReceipt	int = 6
    streamType_global_TransactionState		int = 7

    streamType_count				int = 7
)

type Collector struct {
    apiUrl	string
    history	*history.History

    nodeClient		pb.NodeServiceClient
    meshClient		pb.MeshServiceClient
    globalClient	pb.GlobalStateServiceClient

    streams [streamType_count]bool
    activeStreams int
    online bool

    // Stream status changed.
    notify chan int
}

func NewCollector(nodeAddress string, history *history.History) *Collector {
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
        c.meshClient = pb.NewMeshServiceClient(conn)
        c.globalClient = pb.NewGlobalStateServiceClient(conn)

        err = c.getNetwork()
        if err != nil {
            log.Error("cannot get network info: %s", err)
            time.Sleep(5 * time.Second)
            continue
        }

        go c.syncStatusPump()
        go c.errorPump()
        go c.layersPump()
        go c.accountsPump()
        go c.rewardsPump()
        go c.transactionsReceiptPump()
        go c.transactionsStatePump()

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
            if c.activeStreams == streamType_count {
                c.online = true
            }
            if c.online && c.activeStreams < streamType_count {
                break
            }
        }

        time.Sleep(5 * time.Second)
    }
}
