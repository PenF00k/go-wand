package remgo

import (
	"context"
	"errors"
	"fmt"
	"gitlab.vmassive.ru/wand/goapi/remgo/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"gitlab.vmassive.ru/wand/goapi"
)

//protoc --proto_path=/Users/viktorloskutov/gopath/src/cliwand/proto_client --go_out=plugins=grpc:/Users/viktorloskutov/gopath/src/gitlab.vmassive.ru/wand/goapi/remgo/generated debugwebsocket.proto
//protoc --proto_path=/Users/viktorloskutov/gopath/src/cliwand/proto_client --dart_out=grpc:/Users/viktorloskutov/code/client-pik-remont/client_pik_remont/lib/go_client/proto/generated debugwebsocket.proto

// serveWs handles websocket requests from the peer.
func ServeWs(registry *goapi.Registry, hub *Hub, port int) {
	registry.RegisterEventCallback(hub)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	keepAliveOption := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 5 * time.Minute, // <--- This fixes it!
	})
	log.Tracef("%+v", keepAliveOption)

	grpcServer := grpc.NewServer()
	service := newServer(registry)

	debugwebsocket.RegisterDebugServer(grpcServer, service)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}

type debugServer struct {
	registry *goapi.Registry
}

func newServer(registry *goapi.Registry) *debugServer {
	s := &debugServer{
		registry: registry,
	}
	return s
}

type CallbackWrapper struct {
	gavno   chan []byte
	ssanina chan error
}

func newCallbackWrapper() *CallbackWrapper {
	return &CallbackWrapper{
		gavno:   make(chan []byte, 1),
		ssanina: make(chan error, 1),
	}
}

type eventReceiver struct {
	channel chan *debugwebsocket.SubscriptionEvent
	rwLock  sync.RWMutex
	isAlive bool
}

func (clb *CallbackWrapper) OnSuccess(bytes []byte) {
	log.Printf("method call success")
	clb.gavno <- bytes
}

func (clb *CallbackWrapper) OnError(err string) {
	log.Errorf("CallbackWrapper OnError string: %s", err)
	clb.ssanina <- errors.New(err)

}

func (clb *CallbackWrapper) toSync() (*debugwebsocket.Payload, error) {
	for {
		select {
		case res := <-clb.gavno:
			return &debugwebsocket.Payload{
				Value: res,
			}, nil
		case err := <-clb.ssanina:
			log.Errorf("CallbackWrapper OnError string ssanina: %s", err)
			return nil, err
		}
	}
}

var eventReceivers = make(map[*eventReceiver]bool)

func (d *debugServer) CallMethod(ctx context.Context, args *debugwebsocket.CallMethodArgs) (*debugwebsocket.Payload, error) {
	if args == nil {
		return nil, errors.New("nil args at call method")
	}

	callback := newCallbackWrapper()

	d.registry.Call(args.MethodName, args.Args, callback)
	return callback.toSync()
}

func (d *debugServer) Subscribe(ctx context.Context, args *debugwebsocket.SubscribeArgs) (*debugwebsocket.UnsubscribeArgs, error) {
	fullSubscriptionName, err := d.registry.Subscribe(args.SubscriptionName, args.Args)
	return &debugwebsocket.UnsubscribeArgs{
		FullSubscriptionName: fullSubscriptionName,
	}, err
}

func (d *debugServer) Unsubscribe(ctx context.Context, args *debugwebsocket.UnsubscribeArgs) (*debugwebsocket.Empty, error) {
	d.registry.CancelSubscription(args.FullSubscriptionName)
	return &debugwebsocket.Empty{}, nil
}

func (d *debugServer) RegisterEventCallback(args *debugwebsocket.Empty, server debugwebsocket.Debug_RegisterEventCallbackServer) error {
	receiver := &eventReceiver{
		channel: make(chan *debugwebsocket.SubscriptionEvent),
		isAlive: true,
	}
	log.Tracef("Event receiver created")
	eventReceivers[receiver] = true
	log.Tracef("eventReceivers length = %v", len(eventReceivers))

	defer func() {
		receiver.rwLock.Lock()
		receiver.isAlive = false
		delete(eventReceivers, receiver)
		close(receiver.channel)
		log.Tracef("Event receiver closed: %+v", receiver)
		log.Tracef("eventReceivers = %+v", eventReceivers)
		receiver.rwLock.Unlock()
	}()

	for {
		//dummy := debugwebsocket.SubscriptionEvent{}
		//server.Send(&dummy)

		select {
		case event := <-receiver.channel:
			if err := server.Send(event); err != nil {
				log.Errorf("couldn't send event with error: %v", err)
				return err // we done
			}
		}
	}
}

type Client struct {
	registry *goapi.Registry

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

func (h *Hub) OnEvent(fullSubscriptionName string, bytes []byte) {
	log.Printf("[EVENT] %s", fullSubscriptionName)
	event := &debugwebsocket.SubscriptionEvent{
		FullSubscriptionName: fullSubscriptionName,
		Data:                 bytes,
	}

	for er := range eventReceivers {
		log.Printf("[EVENT] sending to channel %+v", er)
		er.rwLock.RLock()
		if er.isAlive {
			er.channel <- event
		}
		er.rwLock.RUnlock()
	}
}

func NewHub() *Hub {
	hub := &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	log.AddHook(hub)

	return hub
}

func (h *Hub) Levels() []log.Level {
	return log.AllLevels
}

type LogData struct {
	ID string
	// Contains all the fields set by the user.
	Data log.Fields

	// Time at which the log entry was created
	Time time.Time

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level string

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string

	Stack []string
}

type LogEvent struct {
	Log LogData
}

func (h *Hub) Fire(entry *log.Entry) error {
	/*

		type Entry struct {
		Logger *Logger

		// Contains all the fields set by the user.
		Data Fields

		// Time at which the log entry was created
		Time time.Time

		// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
		// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
		Level Level

		// Calling method, with package name
		Caller *runtime.Frame

		// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
		Message string

		// When formatter is called in entry.log(), a Buffer may be set to entry
		Buffer *bytes.Buffer

		// err may contain a field formatting error
		err string
		}
	*/

	event := LogEvent{}
	event.Log.Data = entry.Data
	event.Log.Message = entry.Message
	event.Log.Level = log.Level.String(entry.Level)
	event.Log.Time = entry.Time

	uuid, _ := uuid.NewV4()
	event.Log.ID = uuid.String()

	stack := debug.Stack()
	event.Log.Stack = strings.Split(string(stack), "\n")
	// entry.Caller

	//go h.SendLogEvent(event)

	return nil
}

//func (h *Hub) SendLogEvent(event LogEvent) {
//	resp, _ := json.Marshal(event)
//	h.broadcast <- resp
//}
