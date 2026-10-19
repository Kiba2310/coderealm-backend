package piston

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		primaryURL:  env.GetString("PISTON_LOCAL_URL", "http://localhost:2000/api/v2/execute"),
		fallbackURL: env.GetString("PISTON_CLOUD_URL", "https://emkc.org/api/v2/piston/execute"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type ExecuteRequest struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Files    []File `json:"files"`
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
	reqBody := ExecuteRequest{
		Language: language,
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

type Package struct {
	Language string `json:"language"`
	Version  string `json:"language_version"`
	Installed bool  `json:"installed"`
}

func (c *Client) GetPackages() ([]Package, error) {
	baseURL := strings.TrimSuffix(c.primaryURL, "/execute")
	req, err := http.NewRequest(http.MethodGet, baseURL+"/packages", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("piston API packages returned status: %s", res.Status)
	}

	var packages []Package
	if err := json.NewDecoder(res.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return packages, nil
}

func (c *Client) InstallPackage(language, version string) error {
	payload := map[string]string{
		"language": language,
		"version":  version,
	}
	jsonData, _ := json.Marshal(payload)

	baseURL := strings.TrimSuffix(c.primaryURL, "/execute")
	req, err := http.NewRequest(http.MethodPost, baseURL+"/packages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Temporarily increase timeout for installation as it can be slow
	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to install package, status: %s", res.Status)
	}
	return nil
}

func (c *Client) EnsureRuntime(language string) error {
	if language == "" {
		return nil
	}

	packages, err := c.GetPackages()
	if err != nil {
		fmt.Printf("Failed to get packages from Piston API: %v\n", err)
		return nil // Non-fatal, fallback to attempting to run anyway
	}

	var targetPkg *Package
	lowerLang := strings.ToLower(language)
	for _, pkg := range packages {
		if strings.ToLower(pkg.Language) == lowerLang {
			if targetPkg == nil || pkg.Version > targetPkg.Version {
				p := pkg
				targetPkg = &p
			}
		}
	}

	if targetPkg == nil {
		return fmt.Errorf("language %s not supported by Piston", language)
	}

	if targetPkg.Installed {
		return nil
	}

	fmt.Printf("Installing runtime for %s (v%s)...\n", targetPkg.Language, targetPkg.Version)
	if err := c.InstallPackage(targetPkg.Language, targetPkg.Version); err != nil {
		return fmt.Errorf("failed to install runtime for %s: %w", language, err)
	}
	fmt.Printf("Successfully installed runtime for %s.\n", language)
	return nil
}
