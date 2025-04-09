package api 

import ( 
	"fmt" 
	"log" 
	"strings" 

	"github.com/gin-gonic/gin" 
	"github.com/gorilla/websocket" 

	"github.com/rb4807/Web_Socketing_Real_Time/core" 
) 

func GetAllTickers(c *gin.Context) { 
	c.JSON(200, core.GetAllTickers()) 
} 

func ListenTicker(c *gin.Context) { 
	conn, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024) 
	if err != nil { 
		log.Println("WebSocket Upgrade Error: ", err) 
		return 
	} 
	defer conn.Close() 

	currTicker := c.Param("ticker") 
	log.Println("Current ticker: ", currTicker) 

	if !core.IsTickerAllowed(currTicker) { 
		conn.WriteMessage(websocket.CloseUnsupportedData, []byte("Ticker is not allowed")) 
		log.Println("Ticker not allowed ticker: ", currTicker) 
		return 
	} 

	topic := "trades-" + strings.ToLower(currTicker) 

	// Subscribe to the message bus instead of Kafka
	msgChan := core.Bus.Subscribe(topic)
	
	// Ensure we unsubscribe when the connection is closed
	defer core.Bus.Unsubscribe(topic, msgChan)

	// Handle client disconnection
	conn.SetCloseHandler(func(code int, text string) error { 
		log.Printf("Received connection close request. Closing connection .....") 
		return nil 
	}) 

	// Listen for any messages from the client
	go func() { 
		for {
			_, _, err := conn.ReadMessage()
			if err != nil { 
				log.Println("Error reading message from WS connection. Exiting ...") 
				return 
			}
		}
	}() 

	// Send messages from the bus to the websocket client
	for message := range msgChan { 
		fmt.Println("Sending to client: ", string(message)) 

		err = conn.WriteMessage(websocket.TextMessage, message) 
		if err != nil { 
			log.Println("Error writing message to WS connection: ", err) 
			return 
		} 
	} 
}