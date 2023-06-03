// sky-skills/pkg/clientlib/skillsclient/skillsclient.go
package skillsclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Client represents an HTTP client that can be used to send requests to the skills server.
type Client struct {
	BaseURL    string
	HttpClient *http.Client
	Token      string
	ApiKey     string
}

// Skill represents the structure of a skill.
type Skill struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SkillProject struct {
	SkillID   uuid.UUID `json:"skill_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

func NewClient(baseURL string, token string, apiKey string, httpClient ...*http.Client) *Client {
	var client *http.Client
	if len(httpClient) > 0 {
		client = httpClient[0]
	} else {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &Client{
		BaseURL:    baseURL,
		HttpClient: client,
		Token:      token,
		ApiKey:     apiKey,
	}
}

// CreateSkill sends a POST request to create a new Skill.
func (c *Client) CreateSkill(skill *Skill) (*Skill, error) {
	// TODO: Replace "/skills" with the actual path to the "create skill" endpoint.
	url := fmt.Sprintf("%s/skills", c.BaseURL)

	body, err := json.Marshal(skill)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(bodyBytes))
	}

	var newSkill Skill
	err = json.NewDecoder(resp.Body).Decode(&newSkill)
	if err != nil {
		return nil, err
	}

	return &newSkill, nil
}

// GetSkillByID sends a GET request to retrieve a specific Skill by ID.
func (c *Client) GetSkillByID(id uuid.UUID) (*Skill, error) {
	url := fmt.Sprintf("%s/skills/%s", c.BaseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(bodyBytes))
	}

	var skill Skill
	err = json.NewDecoder(resp.Body).Decode(&skill)
	if err != nil {
		return nil, err
	}

	return &skill, nil
}

// GetAllSkills sends a GET request to retrieve all skills
func (c *Client) GetAllSkills() ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/skills", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var skills []Skill
	if err := json.NewDecoder(resp.Body).Decode(&skills); err != nil {
		return nil, err
	}

	return skills, nil
}

// UpdateSkill sends a PATCH request to update a specific skill
func (c *Client) UpdateSkill(id uuid.UUID, updatedSkill Skill) (Skill, error) {
	updatedSkill.ID = id // Ensure the ID is set correctly
	payload, err := json.Marshal(updatedSkill)
	if err != nil {
		return Skill{}, err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/skills/%s", c.BaseURL, id), bytes.NewBuffer(payload))
	if err != nil {
		return Skill{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return Skill{}, err
	}
	defer resp.Body.Close()

	var skill Skill
	if err := json.NewDecoder(resp.Body).Decode(&skill); err != nil {
		return Skill{}, err
	}

	return skill, nil
}

// DeleteSkill deletes a skill by ID.
func (c *Client) DeleteSkill(skillID uuid.UUID) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/skills/%s", c.BaseURL, skillID), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-KEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete skill: status code %d", resp.StatusCode)
	}

	return nil
}

// SearchSkills searches for skills by a query.
func (c *Client) SearchSkills(query string) ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/skills/search/%s", c.BaseURL, query), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-KEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search skills: status code %d", resp.StatusCode)
	}

	var skills []Skill
	err = json.NewDecoder(resp.Body).Decode(&skills)
	if err != nil {
		return nil, err
	}

	return skills, nil
}

// GetSkillsByCategory retrieves skills by a specific category.
func (c *Client) GetSkillsByCategory(categoryID uuid.UUID) ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/skills/category/%s", c.BaseURL, categoryID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-KEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get skills by category: status code %d", resp.StatusCode)
	}

	var skills []Skill
	err = json.NewDecoder(resp.Body).Decode(&skills)
	if err != nil {
		return nil, err
	}

	return skills, nil
}

func (c *Client) GetSkillsByUserID(userID string) ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/skills/user/%s", c.BaseURL, userID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	var skills []Skill
	err = json.NewDecoder(resp.Body).Decode(&skills)
	if err != nil {
		return nil, err
	}

	return skills, nil
}

func (c *Client) GetPopularSkills(limit int) ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/skills/popular?limit=%s", c.BaseURL, strconv.Itoa(limit)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	var skills []Skill
	err = json.NewDecoder(resp.Body).Decode(&skills)
	if err != nil {
		return nil, err
	}

	return skills, nil
}

func (c *Client) AssociateSkillWithProject(sp SkillProject) error {
	body, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/associate_skill", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) DisassociateSkillFromProject(sp SkillProject) error {
	body, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/disassociate_skill", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) GetProjectIDsForSkill(skillID uuid.UUID) ([]uuid.UUID, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/get_projects", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("skill_id", skillID.String())
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var projectIDs []uuid.UUID
	err = json.NewDecoder(resp.Body).Decode(&projectIDs)
	if err != nil {
		return nil, err
	}

	return projectIDs, nil
}

// GetSkillsForProject sends a GET request to the server to get the skills for a particular project.
func (c *Client) GetSkillsForProject(projectID uuid.UUID) ([]Skill, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/get_skills_for_project", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("project_id", projectID.String())
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var skills []Skill
	err = json.NewDecoder(resp.Body).Decode(&skills)
	if err != nil {
		return nil, err
	}

	return skills, nil
}
