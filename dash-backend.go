package main

import (
    "context"
    "fmt"
    "io"
    "time"
    "flag"

     pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
    "google.golang.org/grpc"
//    "google.golang.org/grpc/grpclog"
    "github.com/golang/protobuf/ptypes/empty"
)

var (
    version string
    commit  string
    branch  string
)

func syncWithNode(nodeClient pb.NodeServiceClient) {
    var protoReq empty.Empty

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := nodeClient.SyncStatusStream(ctx, &protoReq)
    if err != nil {
        fmt.Println("cannot get sync status stream: ", err)
        return
    }

    for {
        res, err := stream.Recv()
        if err == io.EOF {
            return
        }
        if err != nil {
            fmt.Println("cannot receive sync status: ", err)
            return
        }

        fmt.Println("Node sync status: ", res.GetStatus())
    }
}

func main() {
    nodeAddress := flag.String("address", "", "the node address")
    flag.Parse()
    fmt.Println("dial node %s", *nodeAddress)

    conn, err := grpc.Dial(*nodeAddress, grpc.WithInsecure())
    if err != nil {
        fmt.Println("cannot dial node: ", err)
        return
    }

    nodeClient := pb.NewNodeServiceClient(conn)

    syncWithNode(nodeClient)
}
