package remgo

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"gitlab.vmassive.ru/wand/goapi"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	registry *goapi.JsRegistry

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	hub *Hub
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) OnEvent(eventName string, body interface{}) {
	log.Printf("[EVENT] %+#v", body)

	event := EventBody{EventName: eventName, Body: body}
	resp, _ := json.Marshal(event)
	h.broadcast <- resp
}

func (h *Hub) Run(registry *goapi.JsRegistry) {
	registry.RegisterEventCallback(h)

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

type CallRequest struct {
	ID        int
	Call      map[string]interface{}
	Subscribe map[string]interface{}
	Cancel    map[string]interface{}
}

type ResponseBody struct {
	ID      int
	Success interface{}
	Error   interface{}
}

type StatBody struct {
	Name     string
	Request  interface{}
	Response ResponseBody
	Time     time.Duration
}

type EventBody struct {
	EventName string
	Body      interface{}
}

type callMeOnResult struct {
	Time      time.Time
	ID        int
	request   interface{}
	response  chan []byte
	broadcast chan []byte
}

func newRequestHanler(id int, request interface{}, response chan []byte, broadcast chan []byte) *callMeOnResult {
	return &callMeOnResult{
		Time:      time.Now(),
		ID:        id,
		request:   request,
		response:  response,
		broadcast: broadcast,
	}
}

func (call callMeOnResult) SendStat(data ResponseBody) {
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	elapsed := time.Since(call.Time)

	statBody := StatBody{
		Name:     "call",
		Request:  call.request,
		Time:     elapsed,
		Response: data,
	}

	resp, _ := json.Marshal(statBody)
	call.broadcast <- resp
}

func (call callMeOnResult) OnSuccess(data interface{}) {
	log.Printf("[SUCCESS] %+#v", data)

	respBody := ResponseBody{Success: data, ID: call.ID}
	call.SendStat(respBody)

	resp, _ := json.Marshal(respBody)
	call.response <- resp
}

func (call callMeOnResult) OnError(data interface{}) {
	log.Errorf("[ERROR] %+#v", data)

	respBody := ResponseBody{Error: data, ID: call.ID}
	call.SendStat(respBody)

	resp, _ := json.Marshal(respBody)
	call.response <- resp
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		debug.SetPanicOnFault(true)

		request := CallRequest{}
		err := c.conn.ReadJSON(&request)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("message %#+v", request)

		if request.Call != nil {
			log.Printf("calling method %#+v", request)

			callback := newRequestHanler(request.ID, request, c.send, c.hub.broadcast)
			c.registry.Call(request.Call, callback)
		} else if request.Subscribe != nil {
			log.Printf("subscribing %#+v", request)

			c.registry.Subscribe(request.Subscribe)
		} else if request.Cancel != nil {
			log.Errorf("cancelling method %#+v", request)

			c.registry.CancelSubscription(request.Cancel)
		} else {
			log.Errorf("Unknown request %+v", request)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(registry *goapi.JsRegistry, hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		registry: registry,
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
