package main

import (
	"fmt"

	"github.com/joho/godotenv"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

const (
	apiKey string
	apiSecret string
)

func main() {
	err := godotenv.Load();
	if err != nil {
		log.Println("Warning: Could not load .env file.")
	}

	apiKey = os.Getenv("KITE_API_KEY")
	apiSecret = os.Getenv("KITE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("KITE_API_KEY and KITE_API_SECRET must be set in your .env file!")
	}

    kc := kiteconnect.New(apiKey);

	fmt.Println(kc.GetLoginURL());
	var requestToken string
	fmt.Scanf("%s\n", &requestToken)

	data, err := kc.GenerateSession(requestToken, apiSecret)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	kc.SetAccessToken(data.AccessToken)

	margins, err := kc.GetUserMargins()
	if err != nil {
		fmt.Printf("Error getting margins: %v", err)
	}
	fmt.Println("margins: ", margins)

	fmt.Println("Initializing websocket")
	ticker := kiteticker.New(apiKey, data.AccessToken)

	var niftyToken uint32 = 256265

	ticker.OnConnect(func() {
		fmt.Println("Connected to Kite Websocket!")

		err := ticker.Subscribe([]uint32{niftyToken})
		if err != nil {
			fmt.Println("Error setting mode:", err)
		}
	})

	ticker.OnTick(func(tick models.Tick) {
		if tick.InstrumentToken == niftyToken {
			fmt.Printf("\rNIFTY 50 Live Price: ₹%.2f     ", tick.LastPrice)
		}
	})

	ticker.OnError(func(err error) {
		fmt.Println("Error in websocket:", err)
	})

	ticker.OnClose(func(code int, reason string) {
		fmt.Printf("\n WebSocket Closed: %d - %s\n", code, reason)
	})

	ticker.Serve()

}