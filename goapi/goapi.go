package goapi

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(result interface{})
	OnError(err interface{})
}

// JsEvent the interface for any events
type JsEvent interface {
	OnEvent(eventName string, json interface{})
}

type EventCallback interface {
	OnEvent(data interface{})
}

type Subscription interface {
	Cancel()
}

func BuildSubsriptionName(funcName string, params ...interface{}) string {
	list := make([]string, 0, len(params))
	for _, param := range params {
		switch x := param.(type) {
		case string:
			list = append(list, x)
		case int:
			list = append(list, strconv.Itoa(x))

		case int16:
			list = append(list, strconv.FormatInt(int64(x), 10))

		case int32:
			list = append(list, strconv.FormatInt(int64(x), 10))

		case int64:
			list = append(list, strconv.FormatInt(x, 10))

		case float32:
			list = append(list, strconv.FormatFloat((float64)(x), 'f', -1, 64))

		case float64:
			list = append(list, strconv.FormatFloat((float64)(x), 'f', -1, 64))
		}
	}

	tail := strings.Join(list, ":")
	return funcName + ":" + tail
}

// CallFunc type for general callback function
type CallFunc func(map[string]interface{}, JsCallback) error

// SubFunc  type for general event function
type SubFunc func(map[string]interface{}, EventCallback) (Subscription, error)

type JsRegistry struct {
	subscriptions        map[string]SubFunc
	functions            map[string]CallFunc
	subscriptionRegistry SubscriptionRegistry
}

func NewJsRegistry() JsRegistry {
	return JsRegistry{
		subscriptions:        make(map[string]SubFunc),
		functions:            make(map[string]CallFunc),
		subscriptionRegistry: NewSubscriptionRegistry(),
	}
}

func (registry *JsRegistry) RegisterSubscription(eventName string, subFunc SubFunc) {
	registry.subscriptions[eventName] = subFunc
}

func (registry *JsRegistry) RegisterFunction(functionName string, adapterFunction CallFunc) {
	registry.functions[functionName] = adapterFunction
}

func (registry *JsRegistry) RegisterEventCallback(callback JsEvent) {
	registry.subscriptionRegistry.SetCallback(callback)
}

func (registry *JsRegistry) Subsribe(subscriptionData map[string]interface{}) {
	eventName := subscriptionData["event"].(string)
	subscriptionFunc := registry.subscriptions[eventName]
	if subscriptionFunc != nil {
		args, ok := subscriptionData["args"].([]interface{})
		if ok {
			eventName := BuildSubsriptionName(eventName, args)
			registry.subscriptionRegistry.RegisterSubscription(eventName, subscriptionData, subscriptionFunc)
		}
	}
}

func (registry *JsRegistry) CancelSubscription(subscriptionData map[string]interface{}) {
	eventName := subscriptionData["event"].(string)
	subscriptionFunc := registry.subscriptions[eventName]
	if subscriptionFunc != nil {
		args, ok := subscriptionData["args"].([]interface{})
		if ok {
			eventName := BuildSubsriptionName(eventName, args)
			registry.subscriptionRegistry.CancelSubscription(eventName)
		}
	}
}

func (registry *JsRegistry) Call(methodCallData map[string]interface{}, callback JsCallback) {
	methodName := methodCallData["method"].(string)

	functionCall := registry.functions[methodName]
	if functionCall != nil {
		log.Printf(">>>>>>>>>>>>>>>>>>>> calling methodName %s", methodName)
		functionCall(methodCallData, callback)
	} else {
		log.Printf(">>>>>>>>>>>>>>>>>>>> methodName not found %s", methodName)
		callback.OnError("no such method: " + methodName)
	}
}

type JsEventCall struct {
	callback JsEvent
}

func (jsEvent *JsEventCall) SetCallback(callback JsEvent) {
	log.Printf("JsEventCall has new callback")
	jsEvent.callback = callback
}

func (jsEvent JsEventCall) OnEvent(eventName string, data interface{}) {
	if jsEvent.callback != nil {
		log.Printf("sending event %s", eventName)
		jsEvent.callback.OnEvent(eventName, data)
	} else {
		log.Printf("skipping event, no active callbback")
	}
}

type NamedEvent struct {
	eventName string
	callback  *JsEventCall
}

func NewNamedEvent(eventName string, callback *JsEventCall) EventCallback {
	return &NamedEvent{
		eventName: eventName,
		callback:  callback,
	}
}

func (named *NamedEvent) OnEvent(data interface{}) {
	log.Printf("got event %s", named.eventName)
	named.callback.OnEvent(named.eventName, data)
}

type subscriptionData struct {
	counter      int
	subscription Subscription
}

type SubscriptionRegistry struct {
	callback JsEventCall
	active   map[string]*subscriptionData
	lock     sync.Mutex
}

func NewSubscriptionRegistry() SubscriptionRegistry {
	return SubscriptionRegistry{
		active: make(map[string]*subscriptionData),
	}
}

func (registry *SubscriptionRegistry) SetCallback(callback JsEvent) {
	registry.callback.SetCallback(callback)
}

func (registry *SubscriptionRegistry) CancelSubscription(eventName string) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subsription := registry.active[eventName]
	if subsription != nil {
		subsription.counter--
		if subsription.counter == 0 {
			registry.active[eventName] = nil
		}
	}
}

func (registry *SubscriptionRegistry) RegisterSubscription(eventName string, params map[string]interface{}, newCall SubFunc) error {
	log.Printf("new subscriptioin for %s", eventName)
	registry.lock.Lock()
	defer registry.lock.Unlock()

	subsription := registry.active[eventName]
	if subsription == nil {
		event := NewNamedEvent(eventName, &registry.callback)
		sub, err := newCall(params, event)
		if err != nil {
			return err
		}

		subsription = &subscriptionData{
			subscription: sub,
		}

		registry.active[eventName] = subsription
	}

	subsription.counter++

	return nil
}
