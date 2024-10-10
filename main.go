package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

func main() {
	// Load Twitter API credentials from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Could not load the .env file, make sure it exists")
	}

	// Get API keys and tokens from environment variables
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")

	// Set up OAuth1 client configuration
	authConfig := oauth1.NewConfig(consumerKey, consumerSecret)
	userToken := oauth1.NewToken(accessToken, accessSecret)

	// Create an HTTP client for Twitter API requests
	client := authConfig.Client(oauth1.NoContext, userToken)

	// Create a tweet with the current timestamp
	tweetText := fmt.Sprintf("Avinash@Twitter API testing 2!, time now is %s!", time.Now().Format(time.RFC3339))
	tweetID, err := sendTweet(client, tweetText)
	if err != nil {
		log.Fatalf("Error posting tweet: %v", err)
	}
	fmt.Printf("Successfully posted tweet with ID: %s\n", tweetID)

	// Ask the user for tweet ID to delete
	fmt.Println("Enter the tweet ID you want to delete, or press 'Enter' to skip:")
	var tweetToDelete string
	fmt.Scanln(&tweetToDelete)

	// If the user enters an ID, delete the tweet
	if tweetToDelete != "" {
		if err := deleteTweetByID(client, tweetToDelete); err != nil {
			log.Fatalf("Failed to delete tweet: %v", err)
		}
		fmt.Printf("Successfully deleted tweet with ID: %s\n", tweetToDelete)
	} else {
		fmt.Println("No tweet ID entered, exiting.")
	}
}

// sendTweet posts a new tweet and returns the tweet's ID
func sendTweet(client *http.Client, message string) (string, error) {
	apiURL := "https://api.twitter.com/2/tweets"
	tweetData := map[string]string{"text": message}
	jsonData, err := json.Marshal(tweetData)
	if err != nil {
		return "", fmt.Errorf("could not prepare tweet data: %w", err)
	}

	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send tweet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body) // Capture response for debugging
		return "", fmt.Errorf("tweet not created, status: %d, response: %s", resp.StatusCode, string(body))
	}

	// Extract the tweet ID from the response
	var responseBody struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("could not parse response: %w", err)
	}

	return responseBody.Data.ID, nil
}

// deleteTweetByID removes a tweet by its ID
func deleteTweetByID(client *http.Client, tweetID string) error {
	apiURL := fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweetID)

	// Create a DELETE request
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("could not create request to delete tweet: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending delete request: %w", err)
	}
	defer resp.Body.Close()

	// Expect a 200 OK or 204 No Content response
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("could not parse error response: %w", err)
		}
		return fmt.Errorf("failed to delete tweet, status: %d, error: %s", resp.StatusCode, errResp.Error)
	}

	return nil
}
