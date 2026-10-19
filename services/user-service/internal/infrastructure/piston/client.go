package piston

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"BootDevClone/shared/env"
)

type Client struct {
	primaryURL  string
	fallbackURL string
	httpClient  *http.Client
}

func NewClient() *Client {
	return &Client{
		primaryURL:  env.GetString("PISTON_LOCAL_URL", "https://pistonapi.pollak.info/api/v2/execute"),
		fallbackURL: env.GetString("PISTON_CLOUD_URL", "https://emkc.org/api/v2/piston/execute"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type ExecuteRequest struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Files    []File `json:"files"`
	Stdin    string `json:"stdin,omitempty"`
}

type File struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content"`
}

type ExecuteResponse struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Run      Output `json:"run"`
	Compile  Output `json:"compile,omitempty"`
}

type Output struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Output string `json:"output"`
	Code   int    `json:"code"`
	Signal string `json:"signal"`
}

func (c *Client) ExecuteCode(language, code string) (*ExecuteResponse, error) {
	// Normalize language names for Piston
	lang := language
	switch language {
	case "js", "node":
		lang = "javascript"
	case "py":
		lang = "python"
	case "csharp", "c#", "cs":
		lang = "csharp.net"
	case "golang":
		lang = "go"
	}

	reqBody := ExecuteRequest{
		Language: lang,
		Version:  "*",
		Files: []File{
			{Content: code},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Try primary first
	resp, err := c.doRequest(c.primaryURL, jsonData)
	if err != nil {
		// Try fallback if primary fails
		fmt.Printf("Primary Piston API failed (%v), trying fallback...\n", err)
		resp, err = c.doRequest(c.fallbackURL, jsonData)
		if err != nil {
			return nil, fmt.Errorf("both primary and fallback Piston API failed: %w", err)
		}
	}
	return resp, nil
}

func (c *Client) doRequest(url string, data []byte) (*ExecuteResponse, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CodeRealm-Platform/1.0")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("piston API returned status: %s", res.Status)
	}

	var executeResp ExecuteResponse
	if err := json.NewDecoder(res.Body).Decode(&executeResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &executeResp, nil
}
