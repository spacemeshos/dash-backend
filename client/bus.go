package client

// import "github.com/spacemeshos/go-spacemesh/log"

// Hub maintains the set of active clients and broadcasts states to the
// clients.
type Bus struct {
    // Registered clients.
    clients map[*Client]bool

    // State changed.
    Notify chan []byte

    // Register requests from the clients.
    Register chan *Client

    // Unregister requests from clients.
    Unregister chan *Client

    // Last state
    state []byte
}

func NewBus() *Bus {
    return &Bus{
        Notify:     make(chan []byte),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        clients:    make(map[*Client]bool),
        state:      nil,
    }
}

func (bus *Bus) Run() {
    for {
        select {
        case client := <-bus.Register:
            bus.clients[client] = true
            if bus.state != nil {
              client.send <- bus.state
            }
        case client := <-bus.Unregister:
            if _, ok := bus.clients[client]; ok {
                delete(bus.clients, client)
                close(client.send)
            }
        case bus.state = <-bus.Notify:
            // log.Info("Bus: %s", bus.state)
            for client := range bus.clients {
                select {
                case client.send <- bus.state:
                default:
                    close(client.send)
                    delete(bus.clients, client)
                }
            }
        }
    }
}
