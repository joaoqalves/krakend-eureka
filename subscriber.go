package eureka

import (
	"context"
	"github.com/devopsfaith/krakend/sd"
	"github.com/devopsfaith/krakend/config"
	"sync"
	"time"
)

var (
	subscribers               = map[string]sd.Subscriber{}
	subscribersMutex          = &sync.Mutex{}
	fallbackSubscriberFactory = sd.FixedSubscriberFactory
)

func SubscriberFactory(ctx context.Context, c Client) sd.SubscriberFactory {
	return func(cfg *config.Backend) sd.Subscriber {
		if len(cfg.Host) == 0 {
			return fallbackSubscriberFactory(cfg)
		}
		subscribersMutex.Lock()
		defer subscribersMutex.Unlock()
		if sf, ok := subscribers[cfg.Host[0]]; ok {
			return sf
		}
		sf, err := NewSubscriber(ctx, c, cfg.Host[0])
		if err != nil {
			return fallbackSubscriberFactory(cfg)
		}
		subscribers[cfg.Host[0]] = sf
		return sf
	}
}

type Subscriber struct {
	cache  *sd.FixedSubscriber
	mutex  *sync.RWMutex
	client Client
	prefix string
	ctx    context.Context
}

func NewSubscriber(ctx context.Context, c Client, prefix string) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		prefix: prefix,
		cache:  &sd.FixedSubscriber{},
		ctx:    ctx,
		mutex:  &sync.RWMutex{},
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err != nil {
		return nil, err
	}
	*(s.cache) = sd.FixedSubscriber(instances)

	go s.loop()

	return s, nil
}

func (s Subscriber) Hosts() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cache.Hosts()
}

func (s *Subscriber) loop() {

	for {
		time.Sleep(30 * time.Second)
		instances, err := s.client.GetEntries(s.prefix)
		if err != nil {
			continue
		}
		s.mutex.Lock()
		*(s.cache) = sd.FixedSubscriber(instances)
		s.mutex.Unlock()
	}

}

