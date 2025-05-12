package api 

import ( 
	"log" 
	"strings"
	"time"
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin" 
	"github.com/gorilla/websocket" 
	"github.com/google/uuid"

	"github.com/rb4807/Web_Socketing_Real_Time/core" 
) 

// ConsumerInfo provides information about a consumer connection
type ConsumerInfo struct {
	ClientID    string    `json:"client_id"`
	Topic       string    `json:"topic"`
	Connected   time.Time `json:"connected_at"`
	MessagesSent int64    `json:"messages_sent"`
}

var activeConsumers = make(map[string]ConsumerInfo)
var consumersMutex sync.RWMutex

func GetAllTickers(c *gin.Context) { 
	c.JSON(200, core.GetAllTickers()) 
} 

// GetActiveConsumers returns information about all active consumers
func GetActiveConsumers(c *gin.Context) {
	consumers := core.Bus.GetActiveConsumers()
	c.JSON(200, consumers)
}

func ListenTicker(c *gin.Context) { 
	// Generate a unique client ID
	clientID := uuid.New().String()

	conn, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024) 
	if err != nil { 
		log.Println("WebSocket Upgrade Error: ", err) 
		return 
	} 
	defer conn.Close() 

	currTicker := c.Param("ticker") 
	log.Printf("Client %s connected to ticker: %s", clientID, currTicker) 

	if !core.IsTickerAllowed(currTicker) { 
		conn.WriteMessage(websocket.CloseUnsupportedData, []byte("Ticker is not allowed")) 
		log.Println("Ticker not allowed ticker: ", currTicker) 
		return 
	} 

	topic := "trades-" + strings.ToLower(currTicker) 

	// Subscribe to the message bus with client ID
	msgChan := core.Bus.Subscribe(topic, clientID)
	
	// Track this consumer
	consumerInfo := ConsumerInfo{
		ClientID:    clientID,
		Topic:       topic,
		Connected:   time.Now(),
		MessagesSent: 0,
	}
	
	consumersMutex.Lock()
	activeConsumers[clientID] = consumerInfo
	consumersMutex.Unlock()
	
	// Ensure we unsubscribe and clean up when the connection is closed
	defer func() {
		core.Bus.Unsubscribe(topic, clientID)
		
		consumersMutex.Lock()
		delete(activeConsumers, clientID)
		consumersMutex.Unlock()
		
		log.Printf("Client %s disconnected from ticker: %s", clientID, currTicker)
	}()

	// Handle client disconnection
	conn.SetCloseHandler(func(code int, text string) error { 
		log.Printf("Client %s closing connection: %s", clientID, text) 
		return nil 
	}) 

	// Send initial status message
	initialMsg := map[string]interface{}{
		"status": "connected",
		"client_id": clientID,
		"ticker": currTicker,
		"consumer_count": core.Bus.GetConsumerCount(topic),
		"time": time.Now(),
	}
	
	initialJSON, _ := json.Marshal(initialMsg)
	if err := conn.WriteMessage(websocket.TextMessage, initialJSON); err != nil {
		log.Printf("Error sending initial message to client %s: %v", clientID, err)
		return
	}

	// Listen for any messages from the client
	go func() { 
		for {
			_, message, err := conn.ReadMessage()
			if err != nil { 
				log.Printf("Error reading message from client %s: %v", clientID, err) 
				return 
			}
			
			// Echo any client messages back (or process commands)
			log.Printf("Received message from client %s: %s", clientID, string(message))
			
			// Process any client commands here
			// ...
		}
	}() 

	// Send messages from the bus to the websocket client
	for message := range msgChan { 
		// Update message count
		consumersMutex.Lock()
		if info, exists := activeConsumers[clientID]; exists {
			info.MessagesSent++
			activeConsumers[clientID] = info
		}
		consumersMutex.Unlock()
		
		// Send message to client
		if err = conn.WriteMessage(websocket.TextMessage, message); err != nil { 
			log.Printf("Error writing message to client %s: %v", clientID, err) 
			return 
		} 
	} 
}