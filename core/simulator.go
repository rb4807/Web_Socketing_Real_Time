package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"log"
)

// TradeData represents trade information
type TradeData struct {
	Ticker         string    `json:"ticker"`
	Price          float64   `json:"price"`
	Volume         float64   `json:"volume"`
	Timestamp      time.Time `json:"timestamp"`
	ProducerID     string    `json:"producer_id"`
	SequenceNumber int64     `json:"seq_num"`
	ConsumerCount  int       `json:"consumer_count"`
}

var sequenceNumber int64 = 0
var producerID = fmt.Sprintf("producer-%d", time.Now().Unix())

// StartSimulation begins simulating trade data for testing
func StartSimulation() {
	rand.Seed(time.Now().UnixNano())
	
	// Create initial prices for each ticker
	tickerPrices := make(map[string]float64)
	for _, ticker := range GetAllTickers() {
		// Random starting price between 50 and 500
		tickerPrices[ticker] = 50 + rand.Float64()*450
	}
	
	log.Printf("Starting trade simulator with producer ID: %s", producerID)
	
	// Start a goroutine to generate trades
	go func() {
		for {
			// Generate trades for each ticker
			for ticker, basePrice := range tickerPrices {
				// Random price fluctuation
				priceChange := (rand.Float64() - 0.5) * 2 // -1 to +1
				newPrice := basePrice + (priceChange * basePrice * 0.01) // 0-1% change
				
				// Update base price
				tickerPrices[ticker] = newPrice
				
				// Get number of consumers for this ticker
				topic := fmt.Sprintf("trades-%s", strings.ToLower(ticker))
				consumerCount := Bus.GetConsumerCount(topic)
				
				// Only generate data if there are consumers
				if consumerCount > 0 {
					sequenceNumber++
					
					// Create trade data
					trade := TradeData{
						Ticker:         ticker,
						Price:          newPrice,
						Volume:         rand.Float64() * 1000,
						Timestamp:      time.Now(),
						ProducerID:     producerID,
						SequenceNumber: sequenceNumber,
						ConsumerCount:  consumerCount,
					}
					
					// Convert to JSON
					tradeJSON, err := json.Marshal(trade)
					if err != nil {
						fmt.Println("Error marshalling trade data:", err)
						continue
					}
					
					// Publish to message bus and get delivery count
					deliveries := Bus.Publish(topic, tradeJSON)
					
					log.Printf("Published trade #%d for %s to %d/%d consumers", 
						sequenceNumber, ticker, deliveries, consumerCount)
				}
			}
			
			// Sleep for a random interval (100-1000ms)
			time.Sleep(time.Duration(100+rand.Intn(900)) * time.Millisecond)
		}
	}()
}