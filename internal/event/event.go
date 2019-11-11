package event

import (
	"fmt"
	"time"
	"sync"
	"sync/atomic"
)

var (
	NextEventId uint64 = 1
)

var (
	ErrEventBusClosed = fmt.Errorf("event bus was closed")
)

// 参考 https://docs.syncthing.net/dev/events.html

type Event struct {
	// 事件 ID
	ID int64 `json:"id"`

	// 全局 id
	GlobalID int64 `json:"globalID"`

	// "2018-01-03T00:03:29.993038+08:00"
	Time time.Time `json:"time"`

	// "ListenAddressesChanged"
	Type string `json:"type"`

	// 事件内容
	Data interface{} `json:"data"`

	Sender int64 `json:"sender"`

	Reciever int64 `json:"reciever"`

	Processer EventOption
}

// 新建事件
func NewEvent(eventtype string, data interface{}, opts ...EventOption) *Event {
	defer atomic.AddUint64(&NextEventId, uint64(1))
	e := &Event{
		ID: int64(NextEventId),
		GlobalID: int64(NextEventId),
		Time: time.Now(),
		Type: eventtype,
		Data: data,
	}

	for i := range opts {
		opts[i](e)
	}
	return e
}

type EventOption func(*Event)


func WithProcessor(processor EventOption) EventOption {
	return func(evt *Event) {
		evt.Processer = processor
	}
}

// Event 历史记录
type History []Event

// 事件发布者，集成事件总线的使用者可以通过此方法规范监听者的调用。
type EventSource interface {
	// 添加事件监听者
	AddSubscriber(Subscriber)
}

// 事件订阅者
type Subscriber func(event *Event)

// 内部事件总线
type EventBus struct {
	events chan *Event
	done chan struct{}
	group *sync.WaitGroup

	// protect subscribers safe
	subscribersMu sync.RWMutex
	subscribers []Subscriber
}

func NewEventBus() *EventBus {
	eb := &EventBus{
		events: make(chan *Event,1),
		done: make(chan struct{}),
		group: &sync.WaitGroup{},
		subscribers: []Subscriber{},
	}
	eb.group.Add(1)
	go eb.broadcast()
	return eb
}

func (eb *EventBus) Stop() {
	close(eb.done)
	eb.group.Wait()
}

// 订阅事件总线
func (eb *EventBus) Subscribe(ss ...Subscriber) {
	eb.subscribersMu.Lock()
	defer eb.subscribersMu.Unlock()
	for i := range ss {
		eb.subscribers = append(eb.subscribers, ss[i])
	}
}

func (eb *EventBus) Emit(event *Event) error {
	select {
	case eb.events <- event:
		return nil
	case <-eb.done:
		// :-(
		return ErrEventBusClosed
	}
}

// 事件广播使用的是单协程，所以要求各个事件监听者尽快消费事件，
// 不要将耗时的操作放在事件监听中。
func (eb *EventBus) broadcast(/* messages <-chan interface{} */) {
	defer close(eb.events)
	defer eb.group.Done()

	wait := 100 * time.Millisecond
	waitTimer := time.NewTimer(wait)
	defer waitTimer.Stop()

loop:
	for !eb.isDone() {
		select {
		case <-waitTimer.C:
			// 需要定期醒来看看是否结束了
			waitTimer.Reset(wait)
			goto loop
		case evt := <-eb.events:
			eb.subscribersMu.RLock()
			// propagate event to each subscriber
			for i := range eb.subscribers {
				eb.subscribers[i](evt)
			}
			eb.subscribersMu.RUnlock()
			waitTimer.Reset(wait)
		}
	}
}

func (eb *EventBus) isDone() bool {
	select {
	case <-eb.done:
		return true
	default:
		return false
	}
}
