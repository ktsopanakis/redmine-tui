package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Redmine-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

type Issue struct {
	ID          int       `json:"id"`
	Project     Project   `json:"project"`
	Tracker     Tracker   `json:"tracker"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Author      User      `json:"author"`
	AssignedTo  *User     `json:"assigned_to,omitempty"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date,omitempty"`
	DueDate     string    `json:"due_date,omitempty"`
	DoneRatio   int       `json:"done_ratio"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
	Journals    []Journal `json:"journals,omitempty"`
}

type JournalDetail struct {
	Property string `json:"property"`
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

type Journal struct {
	ID        int             `json:"id"`
	User      User            `json:"user"`
	Notes     string          `json:"notes"`
	CreatedOn time.Time       `json:"created_on"`
	Details   []JournalDetail `json:"details"`
}

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Tracker struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Status struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Priority struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type IssuesResponse struct {
	Issues     []Issue `json:"issues"`
	TotalCount int     `json:"total_count"`
	Offset     int     `json:"offset"`
	Limit      int     `json:"limit"`
}

type IssueResponse struct {
	Issue Issue `json:"issue"`
}

type ProjectsResponse struct {
	Projects   []Project `json:"projects"`
	TotalCount int       `json:"total_count"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
}

// GetIssues fetches issues with optional filters
func (c *Client) GetIssues(projectID int, assignedToMe bool, assignedToUserId int, statusOpen bool, limit, offset int) (*IssuesResponse, error) {
	params := url.Values{}
	if projectID > 0 {
		params.Set("project_id", fmt.Sprintf("%d", projectID))
	}
	if assignedToMe {
		params.Set("assigned_to_id", "me")
	} else if assignedToUserId > 0 {
		params.Set("assigned_to_id", fmt.Sprintf("%d", assignedToUserId))
	}
	if statusOpen {
		params.Set("status_id", "open")
	}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	path := fmt.Sprintf("/issues.json?%s", params.Encode())
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response IssuesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetIssue fetches a single issue by ID
func (c *Client) GetIssue(id int) (*Issue, error) {
	path := fmt.Sprintf("/issues/%d.json?include=journals", id)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response IssueResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response.Issue, nil
}

// GetCurrentUser fetches the current user information
func (c *Client) GetCurrentUser() (*User, error) {
	path := "/users/current.json"
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		User User `json:"user"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response.User, nil
}

// GetProjects fetches all projects
func (c *Client) GetProjects(limit, offset int) (*ProjectsResponse, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	path := fmt.Sprintf("/projects.json?%s", params.Encode())
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ProjectsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
