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
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")

	// Create OAuth1 configuration
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	// Create HTTP client
	client := config.Client(oauth1.NoContext, token)

	// Post a new tweet with a unique timestamp
	tweetText := fmt.Sprintf("Avinash Tweeting Test 2!! %s!", time.Now().Format(time.RFC3339))
	tweetID, err := postTweet(client, tweetText)
	if err != nil {
		log.Fatalf("Error posting tweet: %v", err)
	}
	fmt.Printf("Posted tweet with ID: %s\n", tweetID)

	// Wait for user input before deleting the tweet
	var input string
	fmt.Println("Press 'd' to delete the tweet or any other key to exit:")
	fmt.Scanln(&input)

	if input == "d" {
		// Delete the tweet
		if err := deleteTweet(client, tweetID); err != nil {
			log.Fatalf("Error deleting tweet: %v", err)
		}
		fmt.Printf("Deleted tweet with ID: %s\n", tweetID)
	} else {
		fmt.Println("Exiting without deleting the tweet.")
	}
}

// postTweet creates a new tweet and returns the tweet ID
func postTweet(client *http.Client, tweetText string) (string, error) {
	url := "https://api.twitter.com/2/tweets"
	params := map[string]string{"text": tweetText}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tweet data: %w", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to post tweet: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response (201 Created)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body) // Read response body for logging
		return "", fmt.Errorf("failed to post tweet, status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.ID, nil
}

// deleteTweet deletes a tweet by ID
func deleteTweet(client *http.Client, tweetID string) error {
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweetID)

	// Create a new DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete tweet: %w", err)
	}
	defer resp.Body.Close()

	// Accept 200 OK and 204 No Content as successful responses
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("failed to decode error response: %w", err)
		}
		return fmt.Errorf("failed to delete tweet, status code: %d, error: %s", resp.StatusCode, errorResponse.Error)
	}

	return nil
}
