package remgo

//import (
//	"encoding/json"
//	"net/http"
//	"runtime/debug"
//	"strings"
//	"time"
//
//	"github.com/satori/go.uuid"
//
//	"github.com/gorilla/websocket"
//	log "github.com/sirupsen/logrus"
//
//	"gitlab.vmassive.ru/wand/goapi"
//)
//
//var upgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//var (
//	newline = []byte{'\n'}
//	space   = []byte{' '}
//)
//
//const (
//	// Time allowed to write a message to the peer.
//	writeWait = 10 * time.Second
//
//	// Time allowed to read the next pong message from the peer.
//	pongWait = 60 * time.Second
//
//	// Send pings to peer with this period. Must be less than pongWait.
//	pingPeriod = (pongWait * 9) / 10
//
//	// Maximum message size allowed from peer.
//	maxMessageSize = 512
//)
//
//type Client struct {
//	registry *goapi.JsRegistry
//
//	// The websocket connection.
//	conn *websocket.Conn
//
//	// Buffered channel of outbound messages.
//	send chan []byte
//
//	hub *Hub
//}
//
//type Hub struct {
//	// Registered clients.
//	clients map[*Client]bool
//
//	// Inbound messages from the clients.
//	broadcast chan []byte
//
//	// Register requests from the clients.
//	register chan *Client
//
//	// Unregister requests from clients.
//	unregister chan *Client
//}
//
///**
//type Hook interface {
//	Levels() []Level
//	Fire(*Entry) error
//}
//*/
//
//func NewHub() *Hub {
//	hub := &Hub{
//		broadcast:  make(chan []byte),
//		register:   make(chan *Client),
//		unregister: make(chan *Client),
//		clients:    make(map[*Client]bool),
//	}
//
//	log.AddHook(hub)
//
//	return hub
//}
//
//func (h *Hub) Levels() []log.Level {
//	return log.AllLevels
//}
//
//type LogData struct {
//	ID string
//	// Contains all the fields set by the user.
//	Data log.Fields
//
//	// Time at which the log entry was created
//	Time time.Time
//
//	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
//	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
//	Level string
//
//	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
//	Message string
//
//	Stack []string
//}
//
//type LogEvent struct {
//	Log LogData
//}
//
//func (h *Hub) Fire(entry *log.Entry) error {
//	/*
//
//		type Entry struct {
//		Logger *Logger
//
//		// Contains all the fields set by the user.
//		Data Fields
//
//		// Time at which the log entry was created
//		Time time.Time
//
//		// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
//		// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
//		Level Level
//
//		// Calling method, with package name
//		Caller *runtime.Frame
//
//		// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
//		Message string
//
//		// When formatter is called in entry.log(), a Buffer may be set to entry
//		Buffer *bytes.Buffer
//
//		// err may contain a field formatting error
//		err string
//		}
//	*/
//
//	event := LogEvent{}
//	event.Log.Data = entry.Data
//	event.Log.Message = entry.Message
//	event.Log.Level = log.Level.String(entry.Level)
//	event.Log.Time = entry.Time
//
//	uuid, _ := uuid.NewV4()
//	event.Log.ID = uuid.String()
//
//	stack := debug.Stack()
//	event.Log.Stack = strings.Split(string(stack), "\n")
//	// entry.Caller
//
//	go h.SendEvent(event)
//
//	return nil
//}
//
//func (h *Hub) SendEvent(event LogEvent) {
//	resp, _ := json.Marshal(event)
//	h.broadcast <- resp
//}
//
//func (h *Hub) OnEvent(eventName string, body interface{}) {
//	log.Printf("[EVENT] %+#v", body)
//
//	uuid, _ := uuid.NewV4()
//	id := uuid.String()
//
//	event := EventBody{EventName: eventName, Body: body, ID: id}
//	resp, _ := json.Marshal(event)
//	h.broadcast <- resp
//}
//
//func (h *Hub) Run(registry *goapi.JsRegistry) {
//	registry.RegisterEventCallback(h)
//
//	for {
//		select {
//		case client := <-h.register:
//			h.clients[client] = true
//		case client := <-h.unregister:
//			if _, ok := h.clients[client]; ok {
//				delete(h.clients, client)
//				close(client.send)
//			}
//		case message := <-h.broadcast:
//			for client := range h.clients {
//				select {
//				case client.send <- message:
//				default:
//					close(client.send)
//					delete(h.clients, client)
//				}
//			}
//		}
//	}
//}
//
//type CallRequest struct {
//	ID        int
//	Call      map[string]interface{}
//	Subscribe map[string]interface{}
//	Cancel    map[string]interface{}
//}
//
//type ResponseBody struct {
//	ID      int
//	Success interface{}
//	Error   interface{}
//}
//
//type StatBody struct {
//	ID       string
//	Name     string
//	Request  interface{}
//	Response ResponseBody
//	Time     time.Duration
//}
//
//type EventBody struct {
//	ID        string
//	EventName string
//	Body      interface{}
//}
//
//type callMeOnResult struct {
//	Time      time.Time
//	ID        int
//	request   interface{}
//	response  chan []byte
//	broadcast chan []byte
//}
//
//func newRequestHanler(id int, request interface{}, response chan []byte, broadcast chan []byte) *callMeOnResult {
//	return &callMeOnResult{
//		Time:      time.Now(),
//		ID:        id,
//		request:   request,
//		response:  response,
//		broadcast: broadcast,
//	}
//}
//
//func (call callMeOnResult) SendStat(data ResponseBody) {
//	elapsed := time.Since(call.Time)
//
//	uuid, _ := uuid.NewV4()
//	statBody := StatBody{
//		ID:       uuid.String(),
//		Name:     "call",
//		Request:  call.request,
//		Time:     elapsed,
//		Response: data,
//	}
//
//	resp, _ := json.Marshal(statBody)
//	call.broadcast <- resp
//}
//
//func (call callMeOnResult) OnSuccess(data interface{}) {
//	respBody := ResponseBody{Success: data, ID: call.ID}
//	call.SendStat(respBody)
//
//	resp, _ := json.Marshal(respBody)
//	call.response <- resp
//}
//
//func (call callMeOnResult) OnError(data interface{}) {
//	respBody := ResponseBody{Error: data, ID: call.ID}
//	call.SendStat(respBody)
//
//	resp, _ := json.Marshal(respBody)
//	call.response <- resp
//}
//
//// readPump pumps messages from the websocket connection to the hub.
////
//// The application runs readPump in a per-connection goroutine. The application
//// ensures that there is at most one reader on a connection by executing all
//// reads from this goroutine.
//func (c *Client) readPump() {
//	defer func() {
//		c.hub.unregister <- c
//		c.conn.Close()
//	}()
//	c.conn.SetReadLimit(maxMessageSize)
//	c.conn.SetReadDeadline(time.Now().Add(pongWait))
//	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
//	for {
//		debug.SetPanicOnFault(true)
//
//		request := CallRequest{}
//		err := c.conn.ReadJSON(&request)
//		if err != nil {
//			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
//			}
//			break
//		}
//		if request.Call != nil {
//			callback := newRequestHanler(request.ID, request, c.send, c.hub.broadcast)
//			c.registry.Call(request.Call, callback)
//		} else if request.Subscribe != nil {
//			c.registry.Subscribe(request.Subscribe)
//		} else if request.Cancel != nil {
//			c.registry.CancelSubscription(request.Cancel)
//		} else {
//			log.Errorf("Unknown request %+v", request)
//		}
//	}
//}
//
//// writePump pumps messages from the hub to the websocket connection.
////
//// A goroutine running writePump is started for each connection. The
//// application ensures that there is at most one writer to a connection by
//// executing all writes from this goroutine.
//func (c *Client) writePump() {
//	ticker := time.NewTicker(pingPeriod)
//	defer func() {
//		ticker.Stop()
//		c.conn.Close()
//	}()
//	for {
//		select {
//		case message, ok := <-c.send:
//			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if !ok {
//				// The hub closed the channel.
//				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//
//			w, err := c.conn.NextWriter(websocket.TextMessage)
//			if err != nil {
//				return
//			}
//			w.Write(message)
//
//			// Add queued chat messages to the current websocket message.
//			n := len(c.send)
//			for i := 0; i < n; i++ {
//				w.Write(newline)
//				w.Write(<-c.send)
//			}
//
//			if err := w.Close(); err != nil {
//				return
//			}
//		case <-ticker.C:
//			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//				return
//			}
//		}
//	}
//}
//
//// serveWs handles websocket requests from the peer.
//func ServeWs(registry *goapi.JsRegistry, hub *Hub, w http.ResponseWriter, r *http.Request) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	client := &Client{
//		registry: registry,
//		hub:      hub,
//		conn:     conn,
//		send:     make(chan []byte, 256),
//	}
//
//	client.hub.register <- client
//
//	go client.writePump()
//	go client.readPump()
//}
