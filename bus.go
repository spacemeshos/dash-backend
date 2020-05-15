package main

// Hub maintains the set of active clients and broadcasts states to the
// clients.
type Bus struct {
    // Registered clients.
    clients map[*Client]bool

    // State changed.
    notify chan []byte

    // Register requests from the clients.
    register chan *Client

    // Unregister requests from clients.
    unregister chan *Client

    // Last state
    state []byte
}

func newBus() *Bus {
    return &Bus{
        notify:     make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        clients:    make(map[*Client]bool),
        state:      nil,
    }
}

func (bus *Bus) run() {
    for {
        select {
        case client := <-bus.register:
            bus.clients[client] = true
            if bus.state != nil {
              client.send <- bus.state
            }
        case client := <-bus.unregister:
            if _, ok := bus.clients[client]; ok {
                delete(bus.clients, client)
                close(client.send)
            }
        case bus.state = <-bus.notify:
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
