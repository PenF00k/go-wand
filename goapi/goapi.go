package goapi

import (
	"fmt"
	"runtime/debug"
	"strings"
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
	OnEvent(fullSubscriptionName string, bytes []byte)
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

type SubTypesFunc func(args []byte) (string, error)

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

func (registry *Registry) RegisterSubscription(subscriptionName string, subFunc SubFunc, typeFunction SubTypesFunc) {
	registry.subscriptions[subscriptionName] = &subscriptionAdapter{subscriptionFunc: subFunc, subscriptionTypesFunc: typeFunction}
}

func (registry *Registry) RegisterFunction(functionName string, adapterFunction CallFunc) {
	registry.functions[functionName] = adapterFunction
}

func (registry *Registry) RegisterEventCallback(callback Event) {
	registry.subscriptionRegistry.SetCallback(callback)
}

func (registry *Registry) Subscribe(subscriptionName string, args []byte) (string, error) {
	adapter := registry.subscriptions[subscriptionName]
	if adapter == nil {
		return "", fmt.Errorf("no adapter for event %s", subscriptionName)
	}

	fullSubscriptionName, err := adapter.subscriptionTypesFunc(args)
	if err != nil {
		return "", fmt.Errorf("couldn't get Subscribe for event %s with error: %v", subscriptionName, err)
	}

	err = registry.subscriptionRegistry.RegisterSubscription(fullSubscriptionName, args, adapter.subscriptionFunc)
	return fullSubscriptionName, err
}

func (registry *Registry) CancelSubscription(fullSubscriptionName string) {
	subscriptionName := strings.Split(fullSubscriptionName, ":")[0]
	adapter := registry.subscriptions[subscriptionName]
	if adapter != nil {
		registry.subscriptionRegistry.CancelSubscription(fullSubscriptionName)
	} else {
		log.Errorf("No adapter for event %s, fullSubscriptionName: %v", subscriptionName, fullSubscriptionName)
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

func (event EventCall) OnEvent(fullSubscriptionName string, data []byte) {
	if event.callback != nil {
		log.Printf("sending event %s", fullSubscriptionName)
		event.callback.OnEvent(fullSubscriptionName, data)
	} else {
		log.Printf("skipping event, no active callbback")
	}
}

type NamedEvent struct {
	fullSubscriptionName string
	callback             *EventCall
}

func NewNamedEvent(fullSubscriptionName string, callback *EventCall) EventCallback {
	return &NamedEvent{
		fullSubscriptionName: fullSubscriptionName,
		callback:             callback,
	}
}

func (named *NamedEvent) OnEvent(data []byte) {
	log.Printf("got event %s", named.fullSubscriptionName)
	named.callback.OnEvent(named.fullSubscriptionName, data)
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

func (registry *SubscriptionRegistry) CancelSubscription(fullSubscriptionName string) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subscription := registry.active[fullSubscriptionName]
	if subscription != nil {
		subscription.counter--
		log.Infof("subscription for %s was removed", fullSubscriptionName)
		if subscription.counter == 0 {
			delete(registry.active, fullSubscriptionName)
			log.Infof("subscription for %s was completely cancelled", fullSubscriptionName)
			subscription.subscription.Cancel()
		}
	} else {
		log.Errorf("Subscription already cancelled")
	}
}

func (registry *SubscriptionRegistry) RegisterSubscription(fullSubscriptionName string, args []byte, subFunc SubFunc) error {
	log.Infof("new subscription for %s", fullSubscriptionName)
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subscription := registry.active[fullSubscriptionName]
	if subscription == nil {
		event := NewNamedEvent(fullSubscriptionName, &registry.callback)
		sub, err := subFunc(args, event)
		if err != nil {
			return err
		}

		subscription = &subscriptionData{
			subscription: sub,
		}

		registry.active[fullSubscriptionName] = subscription
	}

	subscription.counter++

	return nil
}
