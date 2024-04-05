package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type GPTResponse struct {
	Choices []struct {
		Message string `json:"message"`
	} `json:"choices"`
}

type FightRequest struct {
	Fighter    string `json:"fighter"`
	Challenger string `json:"challenger"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	r := gin.Default()

	// Define your /fight endpoint
	r.POST("/fight", handleFight)

	// Run the server
	r.Run(":" + port)
}

func handleFight(c *gin.Context) {
	var fighters FightRequest
	if err := c.ShouldBindJSON(&fighters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Format"})
		return
	}

	winner := getBattleWinner(fighters.Fighter, fighters.Challenger)
	c.JSON(http.StatusOK, gin.H{"winner": winner})
}

func getBattleWinner(fighter, challenger string) string {
	prompt := "If these two characters fight, who would win? You need to answer only the name of the winner without any other comment\nFighter 1:" + fighter + "\nFighter 2:" + challenger

	winner, err := chatGPTRequest(prompt)
	if err != nil {
		fmt.Println("ERROR: couldn't get the winner with chatgpt, returning fallback")
		fmt.Println(err)
		return fighter
	}
	return winner
}

func chatGPTRequest(prompt string) (string, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not found. Please set the OPENAI_API_KEY environment variable")
	}

	// Create request body
	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.7, // Adjust temperature for response creativity
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	url := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	// Send HTTP request
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response
	var gptResponse GPTResponse
	err = json.Unmarshal(body, &gptResponse)
	if err != nil {
		return "", err
	}

	// Extract and return response text
	if len(gptResponse.Choices) > 0 {
		return gptResponse.Choices[0].Message, nil
	}

	return "", fmt.Errorf("unable to extract response from API")
}
