package goapi

import (
	"fmt"
	"runtime/debug"
	"sync"

	log "github.com/sirupsen/logrus"
)

// FuncCallback the interface for any callbacks
type FuncCallback interface {
	OnSuccess(bytes []byte)
	OnError(err string)
}

// Event the interface for any events
type Event interface {
	OnEvent(eventName string, bytes []byte)
}

type EventCallback interface {
	OnEvent(bytes []byte)
}

type Subscription interface {
	Cancel()
}

// CallFunc type for general callback function
type CallFunc func([]byte, FuncCallback) error

// SubFunc  type for general event function
type SubFunc func([]byte, EventCallback) (Subscription, error)

type SubTypesFunc func(eventName string, callData []byte) (string, error)

type subscriptionAdapter struct {
	subscriptionFunc      SubFunc
	subscriptionTypesFunc SubTypesFunc
}

type Registry struct {
	subscriptions        map[string]*subscriptionAdapter
	functions            map[string]CallFunc
	subscriptionRegistry SubscriptionRegistry
}

func NewRegistry() Registry {
	return Registry{
		subscriptions:        make(map[string]*subscriptionAdapter),
		functions:            make(map[string]CallFunc),
		subscriptionRegistry: NewSubscriptionRegistry(),
	}
}

func (registry *Registry) RegisterSubscription(eventName string, subFunc SubFunc, typeFunction SubTypesFunc) {
	registry.subscriptions[eventName] = &subscriptionAdapter{subscriptionFunc: subFunc, subscriptionTypesFunc: typeFunction}
}

func (registry *Registry) RegisterFunction(functionName string, adapterFunction CallFunc) {
	registry.functions[functionName] = adapterFunction
}

func (registry *Registry) RegisterEventCallback(callback Event) {
	registry.subscriptionRegistry.SetCallback(callback)
}

func (registry *Registry) Subscribe(eventName string, args []byte) {
	adapter := registry.subscriptions[eventName]
	if adapter != nil {
		fullEventName, err := adapter.subscriptionTypesFunc(eventName, args)
		if err == nil {
			err = registry.subscriptionRegistry.RegisterSubscription(fullEventName, args, adapter.subscriptionFunc)
		}
		if err != nil {
			log.Errorf("Couldn't get Subscribe for event %s with error: %v", eventName, err)
		}
	} else {
		log.Errorf("No adapter for event %s", eventName)
	}
}

func (registry *Registry) CancelSubscription(eventName string, args []byte) {
	adapter := registry.subscriptions[eventName]
	if adapter != nil {
		fullEventName, err := adapter.subscriptionTypesFunc(eventName, args)
		if err == nil {
			registry.subscriptionRegistry.CancelSubscription(fullEventName, args)
		} else {
			log.Errorf("Couldn't get CancelSubscription for event %s with error: %v", eventName, err)
		}
	} else {
		log.Errorf("No adapter for event %s", eventName)
	}
}

func (registry *Registry) Call(methodName string, args []byte, callback FuncCallback) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("[!!!] Method \"%s\" crashed", methodName)
			fmt.Printf("%v", r)

			stack := debug.Stack()

			callback.OnError(string(stack))
		}
	}()

	functionCall := registry.functions[methodName]
	if functionCall != nil {
		log.Printf("[CALL] methodName %s", methodName)
		err := functionCall(args, callback)
		if err != nil {
			callback.OnError(err.Error())
		}
	} else {
		log.Errorf("methodName not found %s", methodName)
		callback.OnError("no such method: " + methodName)
	}
}

type EventCall struct {
	callback Event
}

func (event *EventCall) SetCallback(callback Event) {
	event.callback = callback
}

func (event EventCall) OnEvent(eventName string, data []byte) {
	if event.callback != nil {
		log.Printf("sending event %s", eventName)
		event.callback.OnEvent(eventName, data)
	} else {
		log.Printf("skipping event, no active callbback")
	}
}

type NamedEvent struct {
	eventName string
	callback  *EventCall
}

func NewNamedEvent(eventName string, callback *EventCall) EventCallback {
	return &NamedEvent{
		eventName: eventName,
		callback:  callback,
	}
}

func (named *NamedEvent) OnEvent(data []byte) {
	log.Printf("got event %s", named.eventName)
	named.callback.OnEvent(named.eventName, data)
}

type subscriptionData struct {
	counter      int
	subscription Subscription
}

type SubscriptionRegistry struct {
	callback EventCall
	active   map[string]*subscriptionData
	lock     sync.Mutex
}

func NewSubscriptionRegistry() SubscriptionRegistry {
	return SubscriptionRegistry{
		active: make(map[string]*subscriptionData),
	}
}

func (registry *SubscriptionRegistry) SetCallback(callback Event) {
	registry.callback.SetCallback(callback)
}

func (registry *SubscriptionRegistry) CancelSubscription(eventName string, args []byte) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subscription := registry.active[eventName]
	if subscription != nil {
		subscription.counter--
		if subscription.counter == 0 {
			registry.active[eventName] = nil
			subscription.subscription.Cancel()
		}
	} else {
		log.Errorf("Subscription already cancelled")
	}
}

func (registry *SubscriptionRegistry) RegisterSubscription(eventName string, params []byte, newCall SubFunc) error {
	log.Printf("new subscription for %s", eventName)
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subscription := registry.active[eventName]
	if subscription == nil {
		event := NewNamedEvent(eventName, &registry.callback)
		sub, err := newCall(params, event)
		if err != nil {
			return err
		}

		subscription = &subscriptionData{
			subscription: sub,
		}

		registry.active[eventName] = subscription
	}

	subscription.counter++

	return nil
}
