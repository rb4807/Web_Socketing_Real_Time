package core

import (
	"sync"
)

// MessageBus replaces Kafka as the messaging system
type MessageBus struct {
	subscribers map[string][]chan []byte
	mutex       sync.RWMutex
}

var Bus *MessageBus

// InitMessageBus initializes the message bus
func InitMessageBus() {
	Bus = &MessageBus{
		subscribers: make(map[string][]chan []byte),
	}
}

// Subscribe creates a subscription to a specific topic
func (mb *MessageBus) Subscribe(topic string) chan []byte {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	ch := make(chan []byte, 100) // Buffer size of 100
	
	if _, exists := mb.subscribers[topic]; !exists {
		mb.subscribers[topic] = []chan []byte{}
	}
	
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)
	return ch
}

// Unsubscribe removes a subscription from a topic
func (mb *MessageBus) Unsubscribe(topic string, ch chan []byte) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	if _, exists := mb.subscribers[topic]; !exists {
		return
	}
	
	for i, subscriber := range mb.subscribers[topic] {
		if subscriber == ch {
			mb.subscribers[topic] = append(mb.subscribers[topic][:i], mb.subscribers[topic][i+1:]...)
			close(ch)
			break
		}
	}
}

// Publish sends a message to all subscribers of a topic
func (mb *MessageBus) Publish(topic string, message []byte) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	
	if _, exists := mb.subscribers[topic]; !exists {
		return
	}
	
	for _, subscriber := range mb.subscribers[topic] {
		select {
		case subscriber <- message:
			// Message sent successfully
		default:
			// Skip if channel buffer is full
		}
	}
}