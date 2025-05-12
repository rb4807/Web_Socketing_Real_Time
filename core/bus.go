package core

import (
	"sync"
	"time"
)

// MessageBus replaces Kafka as the messaging system
type MessageBus struct {
	subscribers map[string]map[string]chan []byte // topic -> clientID -> channel
	mutex       sync.RWMutex
	stats       map[string]ConsumerStats
	statsMutex  sync.RWMutex
}

// ConsumerStats tracks stats for each consumer
type ConsumerStats struct {
	MessageCount   int64
	LastActiveTime time.Time
	ClientID       string
	Topic          string
}

var Bus *MessageBus

// InitMessageBus initializes the message bus
func InitMessageBus() {
	Bus = &MessageBus{
		subscribers: make(map[string]map[string]chan []byte),
		stats:       make(map[string]ConsumerStats),
	}
}

// Subscribe creates a subscription to a specific topic with a client ID
func (mb *MessageBus) Subscribe(topic string, clientID string) chan []byte {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	ch := make(chan []byte, 100) // Buffer size of 100
	
	if _, exists := mb.subscribers[topic]; !exists {
		mb.subscribers[topic] = make(map[string]chan []byte)
	}
	
	// Store the channel with client ID as key
	mb.subscribers[topic][clientID] = ch
	
	// Initialize stats for this consumer
	mb.statsMutex.Lock()
	mb.stats[clientID+":"+topic] = ConsumerStats{
		MessageCount:   0,
		LastActiveTime: time.Now(),
		ClientID:       clientID,
		Topic:          topic,
	}
	mb.statsMutex.Unlock()
	
	return ch
}

// Unsubscribe removes a subscription from a topic
func (mb *MessageBus) Unsubscribe(topic string, clientID string) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	if _, exists := mb.subscribers[topic]; !exists {
		return
	}
	
	if ch, exists := mb.subscribers[topic][clientID]; exists {
		close(ch)
		delete(mb.subscribers[topic], clientID)
	}
	
	// Clean up stats
	mb.statsMutex.Lock()
	delete(mb.stats, clientID+":"+topic)
	mb.statsMutex.Unlock()
	
	// If no more subscribers for this topic, remove the topic
	if len(mb.subscribers[topic]) == 0 {
		delete(mb.subscribers, topic)
	}
}

// Publish sends a message to all subscribers of a topic
func (mb *MessageBus) Publish(topic string, message []byte) int {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	
	if _, exists := mb.subscribers[topic]; !exists {
		return 0
	}
	
	// Count of successful deliveries
	count := 0
	
	// Send to all subscribers for this topic
	for clientID, ch := range mb.subscribers[topic] {
		select {
		case ch <- message:
			// Message sent successfully
			count++
			
			// Update stats
			mb.statsMutex.Lock()
			if stats, exists := mb.stats[clientID+":"+topic]; exists {
				stats.MessageCount++
				stats.LastActiveTime = time.Now()
				mb.stats[clientID+":"+topic] = stats
			}
			mb.statsMutex.Unlock()
			
		default:
			// Skip if channel buffer is full
		}
	}
	
	return count
}

// GetActiveConsumers returns information about all active consumers
func (mb *MessageBus) GetActiveConsumers() []ConsumerStats {
	mb.statsMutex.RLock()
	defer mb.statsMutex.RUnlock()
	
	consumers := make([]ConsumerStats, 0, len(mb.stats))
	for _, stats := range mb.stats {
		consumers = append(consumers, stats)
	}
	
	return consumers
}

// GetConsumerCount returns the number of consumers for a specific topic
func (mb *MessageBus) GetConsumerCount(topic string) int {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()
	
	if topicSubscribers, exists := mb.subscribers[topic]; exists {
		return len(topicSubscribers)
	}
	
	return 0
}