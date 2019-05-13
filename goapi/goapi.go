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
type CallFunc func([]byte, FuncCallback)

// SubFunc  type for general event function
type SubFunc func([]byte, EventCallback) (Subscription, error)

type SubTypesFunc func(args []byte) (string, error)

type subscriptionAdapter struct {
	subscriptionFunc      SubFunc
	subscriptionTypesFunc SubTypesFunc
}

type Registry struct {
	//subscriptions        map[string]*subscriptionAdapter
	//functions            map[string]CallFunc
	subscriptions        sync.Map
	functions            sync.Map
	subscriptionRegistry SubscriptionRegistry
}

func NewRegistry() Registry {
	return Registry{
		//subscriptions:        make(map[string]*subscriptionAdapter),
		//functions:            make(map[string]CallFunc),
		subscriptions:        sync.Map{},
		functions:            sync.Map{},
		subscriptionRegistry: NewSubscriptionRegistry(),
	}
}

func (registry *Registry) RegisterSubscription(subscriptionName string, subFunc SubFunc, typeFunction SubTypesFunc) {
	//registry.subscriptions[subscriptionName] = &subscriptionAdapter{subscriptionFunc: subFunc, subscriptionTypesFunc: typeFunction}
	registry.subscriptions.Store(subscriptionName, &subscriptionAdapter{subscriptionFunc: subFunc, subscriptionTypesFunc: typeFunction})
}

func (registry *Registry) RegisterFunction(functionName string, adapterFunction CallFunc) {
	log.Infof("Registry RegisterFunction '%v'", functionName)
	//registry.functions[functionName] = adapterFunction
	registry.functions.Store(functionName, adapterFunction)
}

func (registry *Registry) RegisterEventCallback(callback Event) {
	log.Infof("Registry RegisterEventCallback '%v'", callback)
	registry.subscriptionRegistry.SetCallback(callback)
}

func (registry *Registry) Subscribe(subscriptionName string, args []byte) (string, error) {
	log.Infof("Registry Subscribe '%v'", subscriptionName)
	//adapter := registry.subscriptions[subscriptionName]
	value, ok := registry.subscriptions.Load(subscriptionName)
	if !ok || value == nil {
		return "", fmt.Errorf("no adapter for event %s", subscriptionName)
	}

	adapter, ok := value.(*subscriptionAdapter)
	if !ok {
		return "", fmt.Errorf("adapter's type is not *subscriptionAdapter")
	}

	fullSubscriptionName, err := adapter.subscriptionTypesFunc(args)
	if err != nil {
		return "", fmt.Errorf("couldn't get Subscribe for event %s with error: %v", subscriptionName, err)
	}

	err = registry.subscriptionRegistry.RegisterSubscription(fullSubscriptionName, args, adapter.subscriptionFunc)
	return fullSubscriptionName, err
}

func (registry *Registry) CancelSubscription(fullSubscriptionName string) {
	log.Infof("Registry CancelSubscription '%v'", fullSubscriptionName)
	subscriptionName := strings.Split(fullSubscriptionName, ":")[0]

	//adapter := registry.subscriptions[subscriptionName]
	value, ok := registry.subscriptions.Load(subscriptionName)
	if !ok || value == nil {
		log.Errorf("no adapter for event %s", subscriptionName)
		return
	}

	adapter, ok := value.(*subscriptionAdapter)
	if !ok {
		log.Errorf("adapter's type is not *subscriptionAdapter")
		return
	}

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

	//functionCall := registry.functions[methodName]
	value, ok := registry.functions.Load(methodName)
	if !ok || value == nil {
		log.Errorf("no function for method %s", methodName)
		return
	}

	functionCall, ok := value.(CallFunc)
	if !ok {
		log.Errorf("function's type is not CallFunc")
		return
	}

	if functionCall != nil {
		log.Printf("[CALL] methodName %s", methodName)
		//go functionCall(args, callback)
		functionCall(args, callback)
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
		log.Printf("skipping event, no active callback")
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
	//active   map[string]*subscriptionData
	active sync.Map
	lock   sync.Mutex
}

func NewSubscriptionRegistry() SubscriptionRegistry {
	return SubscriptionRegistry{
		//active: make(map[string]*subscriptionData),
		active: sync.Map{},
	}
}

func (registry *SubscriptionRegistry) SetCallback(callback Event) {
	registry.callback.SetCallback(callback)
}

func (registry *SubscriptionRegistry) CancelSubscription(fullSubscriptionName string) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	//subscription := registry.active[fullSubscriptionName]
	value, ok := registry.active.Load(fullSubscriptionName)
	if !ok || value == nil {
		log.Errorf("no subscriptionData for event %s", fullSubscriptionName)
		return
	}

	subscription, ok := value.(*subscriptionData)
	if !ok {
		log.Errorf("subscription's type is not *subscriptionData")
		return
	}

	if subscription != nil {
		subscription.counter--
		log.Infof("subscription for %s was removed", fullSubscriptionName)
		if subscription.counter == 0 {
			//delete(registry.active, fullSubscriptionName)
			registry.active.Delete(fullSubscriptionName)
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

	//subscription := registry.active[fullSubscriptionName]
	var subscription *subscriptionData
	value, ok := registry.active.Load(fullSubscriptionName)
	if ok {
		subscription, ok = value.(*subscriptionData)
	}

	if !ok || subscription == nil {
		event := NewNamedEvent(fullSubscriptionName, &registry.callback)
		sub, err := subFunc(args, event)
		if err != nil {
			return err
		}

		subscription = &subscriptionData{
			subscription: sub,
		}

		//registry.active[fullSubscriptionName] = subscription
		registry.active.Store(fullSubscriptionName, subscription)
	}

	subscription.counter++

	return nil
}
