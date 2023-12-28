package main

import (
	"fmt"
	"sync"
)

type Message struct {
	Topic string
	Data  []byte
}

type Config struct {
	HTTPListenAddr    string
	WSListenAddr      string
	StoreProducerFunc StoreProducerFunc
}

type Server struct {
	*Config
	mu        sync.RWMutex
	topics    map[string]Storer
	peers     map[Peer]bool
	consumers []Consumer
	producers []Producer
	producech chan Message
	quitch    chan struct{}
}

func NewServer(cfg *Config) (*Server, error) {
	producech := make(chan Message)
	s := &Server{
		Config:    cfg,
		topics:    make(map[string]Storer),
		quitch:    make(chan struct{}),
		peers:     make(map[Peer]bool),
		producech: producech,
		producers: []Producer{
			NewHTTPProducer(cfg.HTTPListenAddr, producech),
		},
		consumers: []Consumer{},
	}
	s.consumers = append(s.consumers, NewWSConsumer(cfg.WSListenAddr, s))
	return s, nil
}

func (s *Server) Start() {
	for _, consumer := range s.consumers {
		go func(c Consumer) {
			if err := c.Start(); err != nil {
				fmt.Println(err)
			}
		}(consumer)
	}

	for _, producer := range s.producers {
		go func(p Producer) {
			if err := p.Start(); err != nil {
				fmt.Println(err)
			}
		}(producer)

	}
	s.loop()
}

func (s *Server) loop() {
	for {
		select {
		case <-s.quitch:
			return
		case msg := <-s.producech:
			offset, err := s.publish(msg)
			if err != nil {
				fmt.Println("failed to publish", err)
			} else {
				fmt.Println("produced message", "offset", offset)
			}
		}
	}
}

func (s *Server) publish(msg Message) (int, error) {
	store := s.getStoreForTopic(msg.Topic)
	return store.Push(msg.Data)
}

func (s *Server) getStoreForTopic(topic string) Storer {
	if _, ok := s.topics[topic]; !ok {
		s.topics[topic] = s.StoreProducerFunc()
		fmt.Println("created new topic", topic)
	}
	return s.topics[topic]
}

func (s *Server) AddConn(p Peer) {
	s.mu.Lock()
	defer s.mu.Lock()
	fmt.Println("added new peer")
	s.peers[p] = true
}

func (s *Server) AddPeerToTopics(p Peer, topics ...string) {
	fmt.Println("adding peer to topics", topics, p)
}
