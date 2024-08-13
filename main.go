package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type TotalWorth struct {
	Value float64 `json:"value"`
}

type Balance struct {
	Currency   string    `json:"currency"`
	TotalWorth TotalWorth `json:"totalWorth"`
}

var (
	apiKey     string
	profileID  string
	userToken  string
)

func init() {
	flag.StringVar(&apiKey, "api_key", "", "API key for authorization (can also be set via API_KEY environment variable)")
	flag.StringVar(&profileID, "profile_id", "", "Profile ID to fetch data for (can also be set via PROFILE_ID environment variable)")
	flag.StringVar(&userToken, "token", "", "User token for authentication (can also be set via USER_TOKEN environment variable)")
	flag.Parse()

	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}
	if profileID == "" {
		profileID = os.Getenv("PROFILE_ID")
	}
	if userToken == "" {
		userToken = os.Getenv("USER_TOKEN")
	}

	if apiKey == "" || profileID == "" || userToken == "" {
		log.Fatal("API_KEY, PROFILE_ID, and USER_TOKEN are required either as flags or environment variables")
	}
}

func getFilteredBalances() ([]Balance, error) {
	url := fmt.Sprintf("https://api.wise.com/v4/profiles/%s/balances?types=STANDARD", profileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println("Raw JSON response:", string(body))

	var balances []Balance
	if err := json.Unmarshal(body, &balances); err != nil {
		return nil, err
	}

	return balances, nil
}

func validateTokenFromQuery(r *http.Request) bool {
	token := r.URL.Query().Get("user_token")
	return token == userToken
}

func main() {
	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		if !validateTokenFromQuery(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		balances, err := getFilteredBalances()
		if err != nil {
			http.Error(w, "Failed to fetch or parse data", http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		var filteredResults [][]interface{}
		for _, balance := range balances {
			filteredResults = append(filteredResults, []interface{}{balance.Currency, balance.TotalWorth.Value})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredResults); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		if !validateTokenFromQuery(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		balances, err := getFilteredBalances()
		if err != nil {
			http.Error(w, "Failed to fetch or parse data", http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		var textOutput string
		for _, balance := range balances {
			textOutput += fmt.Sprintf("%s: %.2f\n", balance.Currency, balance.TotalWorth.Value)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(textOutput))
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
