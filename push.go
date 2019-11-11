package push

import (
	"sync"
	"time"
)

type Closer interface {
	Close()
}

type Pusher interface {
	Key() string
	Send() error
	Closer
}

type Receiver interface {
	Key() string
	Receive(PushEvent)
	Closer
}

type PushEvent struct {
	ReceiverKey string
	PusherKey   string
	Time        time.Time
	Data        interface{}
}

type Client struct {
	sync.Mutex
	push chan PushEvent
	done chan struct{}

	pushers   map[string]Pusher
	receivers map[string]Receiver
}

func New() *Client {

	return &Client{
		push:      make(chan PushEvent, 100),
		done:      make(chan struct{}, 1),
		pushers:   make(map[string]Pusher),
		receivers: make(map[string]Receiver),
	}
}

func (c *Client) AddPusher(key string, rc Pusher) {
	c.Lock()
	defer c.Unlock()
	c.pushers[key] = rc
}

func (c *Client) DeletePusher(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.pushers, key)
}

func (c *Client) GetPusher(key string) Pusher {
	c.Lock()
	defer c.Unlock()
	return c.pushers[key]
}

func (c *Client) AddReceiver(key string, rc Receiver) {
	c.Lock()
	defer c.Unlock()
	c.receivers[key] = rc
}

func (c *Client) DeleteReceiver(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.receivers, key)
}

func (c *Client) GetReceiver(key string) Receiver {
	c.Lock()
	defer c.Unlock()
	return c.receivers[key]
}

func (c *Client) Start() {
	for {
		select {
		case e := <-c.push:
			c.Lock()
			defer c.Unlock()
			rc, ok := c.receivers[e.ReceiverKey]
			if ok {
				rc.Receive(e)

			}
			_, ok = c.pushers[e.PusherKey]
			if ok {
				// TODO
			}

		case <-c.done:
			c.Lock()
			defer c.Unlock()
			for _, p := range c.pushers {
				p.Close()
			}
			for _, r := range c.receivers {
				r.Close()
			}
			return
		}
	}
}
