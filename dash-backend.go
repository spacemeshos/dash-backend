package main

import (
    "fmt"
    "net/http"
    "math/rand"
    "os"
    "time"

    "github.com/urfave/cli"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/history"
)

var (
    version string
    commit  string
    branch  string
)

var (
    listenStringFlag      string
    mongoDbUrlStringFlag  string
    mongoDbNameStringFlag string
)

var flags = []cli.Flag{
    cli.StringFlag{
        Name:        "listen",
        Usage:       "Dashboar API listen string in format <host>:<port>",
        Required:    false,
        Destination: &listenStringFlag,
        Value:       ":8080",
    },
    cli.StringFlag{
        Name:        "mongodb",
        Usage:       "Explorer MongoDB Uri string in format mongodb://<host>:<port>",
        Required:    false,
        Destination: &mongoDbUrlStringFlag,
        Value:       "mongodb://localhost:27017",
    },
    cli.StringFlag{
        Name:        "db",
        Usage:       "MongoDB Explorer database name string",
        Required:    false,
        Destination: &mongoDbNameStringFlag,
        Value:       "explorer",
    },
}


func main() {
    app := cli.NewApp()
    app.Name = "Spacemesh Dashboard API Server"
    app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
    app.Flags = flags
    app.Writer = os.Stderr

    app.Action = func(ctx *cli.Context) (error) {

        rand.Seed(time.Now().UTC().UnixNano())

        env, ok := os.LookupEnv("SPACEMESH_API_LISTEN")
        if ok {
            listenStringFlag = env
        }
        env, ok = os.LookupEnv("SPACEMESH_MONGO_URI")
        if ok {
            mongoDbUrlStringFlag = env
        }
        env, ok = os.LookupEnv("SPACEMESH_MONGO_DB")
        if ok {
            mongoDbNameStringFlag = env
        }

        bus := client.NewBus()
        go bus.Run()

        history, err := history.NewHistory(nil, bus, mongoDbUrlStringFlag, mongoDbNameStringFlag)
        if err != nil {
            log.Error("Create History service error: %v", err)
            return err
        }
        go history.Run()

        http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
            client.ServeWs(bus, w, r)
        })

        err = http.ListenAndServe(listenStringFlag, nil)
        if err != nil {
            log.Error("Create HTTP ssrver error: %v", err)
            return err
        }

        log.Info("Server is shutdown")
        return nil
    }

    if err := app.Run(os.Args); err != nil {
        log.Info("%v", err)
        os.Exit(1)
    }

    os.Exit(0)
}

