package synchronizer

import (
    "errors"
    "time"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "google.golang.org/grpc"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/history"
)

const (
    streamType_node_SyncStatus			int = 1
    streamType_node_Error			int = 2
    streamType_mesh_Layer			int = 3
    streamType_globalState			int = 4

    streamType_count				int = 4
)

type Synchronizer struct {
    history	*history.History

    meshClient		pb.MeshServiceClient
    globalClient	pb.GlobalStateServiceClient

    request	chan uint64
}

func NewSynchronizer(history *history.History) *Synchronizer {
    return &Synchronizer{
        history: history,
        request:  make(chan uint64),
    }
}

func (s *Synchronizer) getLayer(uint64 layer) (*types.Layer, error)
{
    // set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req := pb.LayersQueryRequest{ StartLayer: layer, EndLayer: layer }

    response, err := c.meshClient.LayersQuery(ctx, &req)
    if err != nil {
        log.Error("cannot query layer: %v", err)
        return (nil, err)
    }

    if len(response.GetLayer()) != 1 {
        log.Error("wrong result length: %i", len(response.GetLayer()))
        return (nil, errors.New("wrong result length"))
    }

    return (NewLayer(response.GetLayer[0]), nil)
}

func (s *Synchronizer) Run(cc grpc.ClientConnInterface) {
    s.meshClient = pb.NewMeshServiceClient(cc)
    s.globalClient = pb.NewGlobalStateServiceClient(cc)

    for {
        layerNumber := <-s.request

        if layer == uint64(-1) {
            break;
        }

        layer, err := s.getLayer(layerNumber)
        if err != nil {
            continue
        }

        s.history.AddLayer(layer)
    }
}
