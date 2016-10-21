package wskeleton

import "github.com/rpoletaev/container/chain"

const defaultArchiveSize = 30

// var defaultRegisterHandler = func(c *client) {
// 	println("Client added! new client count is ", len h.clients))
// }

// var defaultUnregisterHandler = func(c *client) {

// }

// var defaultBroadcastHandler = func(e *Message) {

// }

type HubConfig struct {
	MaxArchiveSize int
	// 	RegisterHandler   func(c *client)
	// 	UnregisterHandler func(c *client)
	// 	BroadcastHandler  func(e *Event)
}

// func getDefaultConfig() *HubConfig {
// 	return &HubConfig{
// 		MaxArchiveSize:    defaultArchiveSize,
// 		RegisterHandler:   defaultRegisterHandler,
// 		UnregisterHandler: defaultUnregisterHandler,
// 		BroadcastHandler:  defaultBroadcastHandler,
// 	}
// }
type ProcessClient func(c *client)
type ProcessMessage func(msg *Message)

func getDefaultConfig() *HubConfig {
	return &HubConfig{
		MaxArchiveSize: defaultArchiveSize,
	}
}

type Hub struct {
	archive *Archive
	*HubConfig

	clients    map[*client]*struct{}
	broadcast  chan Message
	register   chan *client
	unregister chan *client

	beforeRegister chain.Chain
	afterRegister  chain.Chain

	beforeBroadcast chain.Chain
	afterBroadcast  chain.Chain

	beforeUnregister chain.Chain
	afterUnregister  chain.Chain
}

func CreateHub(config *HubConfig) *Hub {
	if config == nil {
		config = getDefaultConfig()
	} else {
		if config.MaxArchiveSize <= 0 {
			config.MaxArchiveSize = defaultArchiveSize
		}
	}

	return &Hub{
		HubConfig:        config,
		broadcast:        make(chan Message),
		register:         make(chan *client),
		unregister:       make(chan *client),
		clients:          make(map[*client]*struct{}),
		archive:          CreateArchive(config.MaxArchiveSize),
		beforeRegister:   chain.Chain{},
		afterRegister:    chain.Chain{},
		beforeBroadcast:  chain.Chain{},
		afterBroadcast:   chain.Chain{},
		beforeUnregister: chain.Chain{},
		afterUnregister:  chain.Chain{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			runProcessClientChain(h.beforeRegister, c)
			h.clients[c] = nil
			h.sendArchive(c)
			runProcessClientChain(h.afterRegister, c)
			break

		case c := <-h.unregister:
			_, ok := h.clients[c]
			if ok {
				runProcessClientChain(h.beforeUnregister, c)
				delete(h.clients, c)
				close(c.send)
				runProcessClientChain(h.afterUnregister, c)
			}
			break

		case m := <-h.broadcast:
			runProcessMessageChain(h.beforeBroadcast, &m)
			h.archive.Add(m)
			h.broadcastMessage(&m)
			runProcessMessageChain(h.afterBroadcast, &m)
			break
		}
	}
}

func (h *Hub) broadcastMessage(msg *Message) {
	for c := range h.clients {
		select {
		case c.send <- *msg:
			break

		default:
			close(c.send)
			delete(h.clients, c)
		}
	}
}

func (h *Hub) sendArchive(c *client) {
	h.archive.Each(func(message Message) {
		c.send <- message
	})
}

func (h *Hub) RegisterClient(c *client) {
	h.register <- c
}

func (h *Hub) Unregister(c *client) {
	h.unregister <- c
}

func (h *Hub) AddBeforeRegister(pc ProcessClient) {
	h.beforeRegister.Add(pc)
}

func (h *Hub) AddAfterRegister(pc ProcessClient) {
	h.afterRegister.Add(pc)
}

func (h *Hub) AddBeforeUnregister(pc ProcessClient) {
	h.beforeUnregister.Add(pc)
}

func (h *Hub) AddAfterUnregister(pc ProcessClient) {
	h.afterUnregister.Add(pc)
}

func (h *Hub) AddBeforeBroadcast(pm ProcessMessage) {
	h.beforeBroadcast.Add(pm)
}

func (h *Hub) AddAfterBroadcast(pm ProcessMessage) {
	h.afterBroadcast.Add(pm)
}

type Message struct {
	Type string      `json:"event"`
	Data interface{} `json:"data"`
}

func runProcessClientChain(c chain.Chain, cli *client) {
	c.Each(func(v interface{}) {
		v.(ProcessClient)(cli)
	})
}

func runProcessMessageChain(c chain.Chain, msg *Message) {
	c.Each(func(v interface{}) {
		v.(ProcessMessage)(msg)
	})
}

func (h *Hub) SetArchive(arc *Archive) {
	h.archive = arc
}
