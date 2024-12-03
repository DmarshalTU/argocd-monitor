package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"crypto/tls"

	"argocd-monitor/internal/models"
)

type ArgocdService struct {
	baseURL       string
	client        *http.Client
	adminPassword string
}

func NewArgocdService() *ArgocdService {
	// Create HTTP client that skips TLS verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	password := os.Getenv("ARGOCD_ADMIN_PASSWORD")
	if password == "" {
		log.Println("Warning: ARGOCD_ADMIN_PASSWORD environment variable not set")
	}

	return &ArgocdService{
		baseURL:       "https://argo-cd-1733240091-server.default.svc.cluster.local/api/v1",
		client:        client,
		adminPassword: password,
	}
}

func (s *ArgocdService) login() (string, error) {
	loginURL := fmt.Sprintf("%s/session", s.baseURL)
	payload := fmt.Sprintf(`{"username":"admin","password":"%s"}`, s.adminPassword)
	
	log.Printf("Attempting to login to ArgoCD at %s", loginURL)
	
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute login request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}
	
	var result struct {
		Token string `json:"token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode login response: %v", err)
	}
	
	log.Println("Successfully logged in to ArgoCD")
	return result.Token, nil
}

func (s *ArgocdService) GetApplications() ([]models.Application, error) {
	token, err := s.login()
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	appsURL := fmt.Sprintf("%s/applications", s.baseURL)
	log.Printf("Fetching applications from %s", appsURL)

	req, err := http.NewRequest("GET", appsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create applications request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute applications request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("applications request failed with status code: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Sync struct {
					Status string `json:"status"`
				} `json:"sync"`
				Health struct {
					Status string `json:"status"`
				} `json:"health"`
			} `json:"status"`
			Spec struct {
				Destination struct {
					Namespace string `json:"namespace"`
				} `json:"destination"`
				Project string `json:"project"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode applications response: %v", err)
	}

	var apps []models.Application
	for _, item := range result.Items {
		apps = append(apps, models.Application{
			Name:       item.Metadata.Name,
			SyncStatus: item.Status.Sync.Status,
			Health:     item.Status.Health.Status,
			Namespace:  item.Spec.Destination.Namespace,
			Project:    item.Spec.Project,
		})
	}

	log.Printf("Successfully fetched %d applications", len(apps))
	return apps, nil
}