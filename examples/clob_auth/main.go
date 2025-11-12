package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	apiKeyStr := os.Getenv("API_KEY")
	if apiKeyStr == "" {
		log.Fatal("API_KEY not set in environment")
	}

	apiSecretyStr := os.Getenv("API_SECRET")
	if apiSecretyStr == "" {
		log.Fatal("API_SECRET not set in environment")
	}

	apiPassStr := os.Getenv("API_PASSPHRASE")
	if apiPassStr == "" {
		log.Fatal("API_PASSPHRASE not set in environment")
	}

	// Create a new client with options
	client := polymarketdataclient.New(
		// polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithLogger(polymarketdataclient.NewSilentLogger()),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("Connected to Polymarket WebSocket!")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			log.Printf("Received raw message: %s\n", string(data))
			var msg polymarketdataclient.Message
			err := json.Unmarshal(data, &msg)
			if err != nil {
				log.Printf("Invalid message %v received: %v", msg, err)
				return
			}

			// fmt.Printf("Received msg: %+v\n", msg)

			switch msg.Topic {
			case polymarketdataclient.TopicClobUser:
				switch msg.Type {
				case polymarketdataclient.MessageTypeTrade:
					var trade polymarketdataclient.CLOBTrade
					err = json.Unmarshal(msg.Payload, &trade)
					if err != nil {
						log.Printf("Invalid trade %v received: %v", msg.Payload, err)
						return
					}

					log.Printf("Trade: %+v\n", trade)
				}
			case polymarketdataclient.TopicComments:
				// TODO:
			}
		}),
	)

	// Connect to the server
	if err := client.Connect(); err != nil {
		panic(err)
	}

	// Subscribe to market data
	subscriptions := []polymarketdataclient.Subscription{
		{
			Topic: polymarketdataclient.TopicClobUser,
			Type:  polymarketdataclient.MessageTypeAll,
			ClobAuth: &polymarketdataclient.ClobAuth{
				Key:        apiKeyStr,
				Secret:     apiSecretyStr,
				Passphrase: apiPassStr,
			},
		},
	}

	if err := client.Subscribe(subscriptions); err != nil {
		panic(err)
	}

	// Keep the connection alive
	time.Sleep(30 * time.Second)

	// Clean up
	client.Disconnect()
}
