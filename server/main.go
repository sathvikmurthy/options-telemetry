package main

import (
	"fmt"
	"encoding/json"
	"log"
	"os"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	// "github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

var (
	apiKey string
	apiSecret string
	kc *kiteconnect.Client
	ticker *kiteticker.Ticker
)

var upgrader = websocket.Upgrader {
	CheckOrigin: func(r *http.Request) bool { return true },
}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan TickData)

type TickData struct {
	InstrumentToken uint32  `json:"instrument_token"`
	LastPrice       float64 `json:"last_price"`
}

func main() {
	godotenv.Load()
	apiKey = os.Getenv("KITE_API_KEY")
	apiSecret = os.Getenv("KITE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("KITE_API_KEY and KITE_API_SECRET must be set in your .env file!")
	}

    kc = kiteconnect.New(apiKey);

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login-url", getLoginURL)
	mux.HandleFunc("/api/start-session", startSession)
	mux.HandleFunc("/ws", handleLocalWS)

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

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func startTicker(accessToken string) {
	ticker = kiteticker.New(apiKey, accessToken)

	ticker.OnConnect(func() {
		fmt.Println("Connected to Kite Ticker! Subscribing to NIFTY 50...")

		ticker.Subscribe([]uint32{256265})
		ticker.SetMode(kiteticker.ModeFull, []uint32{256265})
	})

	ticker.OnTick(func(tick models.Tick) {
		broadcast <- TickData{
			InstrumentToken: tick.InstrumentToken,
			LastPrice:       tick.LastPrice,
		}
	})

	ticker.OnError(func(err error) { fmt.Println("Ticker Error:", err) })
	ticker.OnClose(func(code int, reason string) { fmt.Println("Ticker Closed:", code, reason) })

	go ticker.Serve()
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

//CORS to talk to the frontend
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