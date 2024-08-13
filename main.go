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

// Nested struct to handle the totalWorth field if it's an object
type TotalWorth struct {
	Value float64 `json:"value"`
}

// Balance represents the structure of the balance data we're interested in
type Balance struct {
	Currency   string    `json:"currency"`
	TotalWorth TotalWorth `json:"totalWorth"`
}

func main() {
	// Define command-line flags
	apiKey := flag.String("api_key", "", "API key for authorization (can also be set via API_KEY environment variable)")
	profileID := flag.String("profile_id", "", "Profile ID to fetch data for (can also be set via PROFILE_ID environment variable)")

	// Parse command-line flags
	flag.Parse()

	// Use environment variables if flags are not provided
	if *apiKey == "" {
		*apiKey = os.Getenv("API_KEY")
	}
	if *profileID == "" {
		*profileID = os.Getenv("PROFILE_ID")
	}

	// Ensure both API_KEY and PROFILE_ID are provided
	if *apiKey == "" || *profileID == "" {
		log.Fatal("API_KEY and PROFILE_ID are required either as flags or environment variables")
	}

	// Function to fetch and filter the balances
	getFilteredBalances := func() ([]Balance, error) {
		// Define the URL with the PROFILE_ID and filter parameters
		url := fmt.Sprintf("https://api.wise.com/v4/profiles/%s/balances?types=STANDARD", *profileID)

		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Set the Authorization header with the Bearer token
		req.Header.Set("Authorization", "Bearer "+*apiKey)

		// Perform the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Print the raw JSON for debugging purposes
		log.Println("Raw JSON response:", string(body))

		// Parse the JSON data
		var balances []Balance
		if err := json.Unmarshal(body, &balances); err != nil {
			return nil, err
		}

		return balances, nil
	}

	// Endpoint to return raw JSON data after filtering
	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		balances, err := getFilteredBalances()
		if err != nil {
			http.Error(w, "Failed to fetch or parse data", http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		// Create a slice to hold the filtered results
		var filteredResults [][]interface{}
		for _, balance := range balances {
			filteredResults = append(filteredResults, []interface{}{balance.Currency, balance.TotalWorth.Value})
		}

		// Return the filtered JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredResults); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})

	// Endpoint to return a text representation of the filtered data
	http.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		balances, err := getFilteredBalances()
		if err != nil {
			http.Error(w, "Failed to fetch or parse data", http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		// Create a textual representation of the filtered results
		var textOutput string
		for _, balance := range balances {
			textOutput += fmt.Sprintf("%s: %.2f\n", balance.Currency, balance.TotalWorth.Value)
		}

		// Return the text response
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(textOutput))
	})

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
