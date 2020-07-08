package collector

import (
    "time"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "google.golang.org/grpc"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/history"
)

const (
    streamType_node_SyncStatus			int = 1
    streamType_mesh_Layer			int = 2
    streamType_globalState			int = 3
    streamType_node_Error			int = 4

    streamType_count				int = 3
)

type Collector struct {
    apiUrl	string
    history	*history.History

    nodeClient		pb.NodeServiceClient
    meshClient		pb.MeshServiceClient
    globalClient	pb.GlobalStateServiceClient

    streams [streamType_count]bool
    activeStreams int
    connecting bool
    online bool
    closing bool

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
        log.Info("dial node %v", c.apiUrl)
        c.connecting = true

        conn, err := grpc.Dial(c.apiUrl, grpc.WithInsecure())
        if err != nil {
            log.Error("cannot dial node: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        c.nodeClient = pb.NewNodeServiceClient(conn)
        c.meshClient = pb.NewMeshServiceClient(conn)
        c.globalClient = pb.NewGlobalStateServiceClient(conn)

        err = c.getNetworkInfo()
        if err != nil {
            log.Error("cannot get network info: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        go c.syncStatusPump()
//        go c.errorPump()
        go c.layersPump()
        go c.globalStatePump()

        for ; c.connecting || c.closing || c.online; {
            state := <-c.notify
            switch {
            case state > 0:
                c.streams[state - 1] = true
                c.activeStreams++
                log.Info("stream connected %v", state)
            case state < 0:
                c.streams[(-state) - 1] = false
                c.activeStreams--
                if c.activeStreams == 0 {
                    c.closing = false
                }
                log.Info("stream disconnected %v", state)
            }
            if c.activeStreams == streamType_count {
                c.connecting = false
                c.online = true
                log.Info("all streams synchronized!")
            }
            if c.online && c.activeStreams < streamType_count {
                log.Info("streams desynchronized!!!")
                c.online = false
                c.closing = true
                conn.Close()
            }
        }

        log.Info("Wait 5 seconds...")
        time.Sleep(5 * time.Second)
    }
}
