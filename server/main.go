package main

import (
	"fmt"
	"encoding/json"
	"log"
	"os"
	"net/http"

	"github.com/joho/godotenv"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	// "github.com/zerodha/gokiteconnect/v4/models"
	// kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

var (
	apiKey string
	apiSecret string
	kc *kiteconnect.Client
)

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