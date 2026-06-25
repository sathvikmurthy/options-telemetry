package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

var (
	apiKey    string
	apiSecret string
	kc        *kiteconnect.Client
	ticker    *kiteticker.Ticker
)

// selected by the user from getPositions
var activeShortToken uint32
var activeLongToken uint32

type SpreadData struct {
	NiftyLTP      float64 `json:"niftyLTP"`
	ShortLTP      float64 `json:"shortLTP"`
	LongLTP       float64 `json:"longLTP"`
	NetSpread     float64 `json:"netSpread"`
	InitialSpread float64 `json:"initialSpread"`
	Status        string  `json:"status"`
}

var currentData SpreadData

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan SpreadData)

func main() {
	godotenv.Load()
	apiKey = os.Getenv("KITE_API_KEY")
	apiSecret = os.Getenv("KITE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("KITE_API_KEY and KITE_API_SECRET must be set in your .env file!")
	}

	kc = kiteconnect.New(apiKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login-url", getLoginURL)
	mux.HandleFunc("/api/start-session", startSession)
	mux.HandleFunc("/ws", handleLocalWS)
	mux.HandleFunc("/api/positions", getPositions)
	mux.HandleFunc("/api/track-spread", setTrackedSpread)

	go handleMessages()

	fmt.Println("Auth Backend Running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}

func getLoginURL(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"url": kc.GetLoginURL()})
}

func startSession(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"request_token"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	data, err := kc.GenerateSession(request.Token, apiSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kc.SetAccessToken(data.AccessToken)
	fmt.Println("✅ Session Generated Successfully! Access Token created.")

	startTicker(data.AccessToken)

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func getPositions(w http.ResponseWriter, r *http.Request) {
	if kc == nil {
		http.Error(w, "Kite client not initialized. Please log in first.", http.StatusInternalServerError)
		return
	}

	positions, err := kc.GetPositions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func setTrackedSpread(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShortToken uint32 `json:"short_token"`
		LongToken  uint32 `json:"long_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	activeShortToken = req.ShortToken
	activeLongToken = req.LongToken

	positions, err := kc.GetPositions()
	if err == nil {
		var shortEntry, longEntry float64

		for _, pos := range positions.Net {
			if pos.InstrumentToken == activeShortToken {
				shortEntry = pos.AveragePrice
			}
			if pos.InstrumentToken == activeLongToken {
				longEntry = pos.AveragePrice
			}
		}

		if shortEntry > 0 && longEntry > 0 {
			currentData.InitialSpread = shortEntry - longEntry
		}
	} else {
		fmt.Println("Warning: Could not fetch positions to calculate initial spread:", err)
	}

	if ticker != nil {
		ticker.Subscribe([]uint32{activeShortToken, activeLongToken})
		ticker.SetMode(kiteticker.ModeFull, []uint32{activeShortToken, activeLongToken})
		currentData.Status = "Spread Tracked - Waiting for ticks..."
		fmt.Printf("Tracking Spread: Short [%d] | Long [%d]\n", activeShortToken, activeLongToken)
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func startTicker(accessToken string) {
	ticker = kiteticker.New(apiKey, accessToken)

	ticker.OnConnect(func() {
		fmt.Println("Connected to Kite Ticker! Subscribing to NIFTY 50...")

		ticker.Subscribe([]uint32{256265})
		ticker.SetMode(kiteticker.ModeFull, []uint32{256265})
	})

	ticker.OnTick(func(tick models.Tick) {
		//nifty
		if tick.InstrumentToken == 256265 {
			currentData.NiftyLTP = tick.LastPrice
		}

		// dynamic short leg
		if activeShortToken != 0 && tick.InstrumentToken == activeShortToken {
			currentData.ShortLTP = tick.LastPrice
		}

		// dynamic long leg
		if activeLongToken != 0 && tick.InstrumentToken == activeLongToken {
			currentData.LongLTP = tick.LastPrice
		}

		// calculate net spread
		if currentData.ShortLTP > 0 && currentData.LongLTP > 0 {
			currentData.NetSpread = currentData.ShortLTP - currentData.LongLTP

			// update status
			if currentData.Status == "Spread Tracked - Waiting for ticks..." {
				currentData.Status = "Monitoring Live Spread"
			}

			targetPrice := currentData.InitialSpread / 2

			// target
			if currentData.NetSpread <= targetPrice && currentData.InitialSpread > 0 {
				currentData.Status = "Target Reached - Executing Close!"
			}
		}

		broadcast <- currentData
	})

	go ticker.Serve()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allowing all origins for dev
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleLocalWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	clients[ws] = true

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			delete(clients, ws)
			break
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// CORS to talk to the frontend
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
