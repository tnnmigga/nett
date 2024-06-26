package eventbus

import (
	"fmt"
	"strconv"

	"github.com/tnnmigga/core/idef"
	"github.com/tnnmigga/core/infra/zlog"
	"github.com/tnnmigga/core/msgbus"
	"github.com/tnnmigga/core/utils"
)

type ISubscriber interface {
	Name() string
	Topics() []string
	Handler(event *Event)
}

type EventBus struct {
	subs map[string][]ISubscriber
}

func (bus *EventBus) RegisterSubscriber(sub ISubscriber) {
	if bus.find(sub) {
		zlog.Errorf("%s has registered", sub.Name())
		return
	}
	for _, topic := range sub.Topics() {
		bus.subs[topic] = append(bus.subs[sub.Name()], sub)
	}
}

func (bus *EventBus) RegisterHandler(topic string, handler func(event *Event)) {
	h := &eventHandler{
		name:    utils.FuncName(handler),
		topic:   topic,
		handler: handler,
	}
	bus.RegisterSubscriber(h)
}

func (bus *EventBus) UnregisterSubscriber(sub ISubscriber) {
	bus.removeSubscriber(sub.Name())
}

func (bus *EventBus) UnregisterHandler(topic string, handler func(event *Event)) {
	bus.removeSubscriber(utils.FuncName(handler))
}

func (bus *EventBus) removeSubscriber(name string) {
	for topic, subs := range bus.subs {
		bus.subs[topic] = utils.Filter(subs, func(sub ISubscriber) bool {
			return sub.Name() != name
		})
	}
}

func (bus *EventBus) find(sub ISubscriber) bool {
	subName := sub.Name()
	for _, sub := range bus.subs {
		for _, v := range sub {
			if v.Name() == subName {
				return true
			}
		}
	}
	return false
}

func (bus *EventBus) dispatch(event *Event) {
	subs := bus.subs[event.Topic]
	for _, sub := range subs {
		utils.ExecAndRecover(func() {
			sub.Handler(event)
		})
	}
}

func New(m idef.IModule) *EventBus {
	bus := &EventBus{
		subs: map[string][]ISubscriber{},
	}
	msgbus.RegisterHandler(m, bus.dispatch)
	return bus
}

func (bus *EventBus) Cast(event *Event) {
	msgbus.Cast(event)
}

func (bus *EventBus) SyncCast(event *Event) {
	bus.dispatch(event)
}

type eventHandler struct {
	name    string
	topic   string
	handler func(*Event)
}

func (h *eventHandler) Name() string {
	return h.name
}

func (h *eventHandler) Topics() []string {
	return []string{h.topic}
}

func (h *eventHandler) Handler(event *Event) {
	h.handler(event)
}

func (e *Event) Int(name string) int {
	if n, err := strconv.Atoi(e.Str(name)); err == nil {
		return n
	}
	panic(fmt.Errorf("event param %s not a number ", name))
}

func (e *Event) Str(name string) (arg string) {
	if e.Args != nil {
		if v, ok := e.Args[name]; ok {
			return v
		}
	}
	panic(fmt.Errorf("event param %s not found", name))
}
