package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	url   string
	key   string
	resty *resty.Client

	Config *ConfigService
}
type ConfigService struct{ client *Client }

func New(url string, key string) *Client {
	return NewWithClient(&http.Client{Timeout: 10 * time.Second}, url, key)
}

func NewWithClient(c *http.Client, url string, key string) *Client {
	client := &Client{
		url,
		key,
		resty.NewWithClient(c),

		nil,
	}

	client.Config = &ConfigService{client}

	return client
}

type response struct {
	Success bool
	Data    interface{}
	Error   *string
}

// Posts a raw request with `payload` to `endpoint`.
func (c *Client) Request(endpoint string, payload map[string]interface{}) (interface{}, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("Failed to marshal request payload.")
	}

	resp, err := c.resty.R().
		SetFormData(map[string]string{
			"key":  c.key,
			"data": string(data),
		}).
		Post(c.url + "/" + endpoint)
	if err != nil {
		return nil, err
	}

	r := new(response)
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return nil, err
	}

	// Handle errors from the API
	if r.Error != nil {
		return nil, errors.New(*r.Error)
	}

	return r.Data, err
}

// Returns the full configuration tree at the specified path
func (svc *ConfigService) ShowTree(path string) (map[string]interface{}, error) {
	resp, err := svc.client.Request("retrieve", map[string]interface{}{
		"op":   "showConfig",
		"path": strings.Split(path, " "),
	})
	if err != nil {
		if strings.Contains(err.Error(), "specified path is empty") {
			// If we get an empty path error, consume it and return nil
			return nil, nil
		} else {
			return nil, err
		}
	}

	obj, ok := resp.(map[string]interface{})
	if !ok {
		return nil, errors.New("Received unexpected repsonse format from server.")
	}

	return obj, nil
}

// Returns the single configuration value at the speicfied path
func (svc *ConfigService) Show(path string) (*string, error) {
	obj, err := svc.ShowTree(path)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}

	components := strings.Split(path, " ")
	terminal := components[len(components)-1]

	val, ok := obj[terminal].(string)
	if !ok {
		return nil, errors.New("Value missing from configuration tree returned by server.")
	}

	return &val, nil
}

// Sets a configuration value at the specified path
func (svc *ConfigService) Set(path string, value string) error {
	_, err := svc.client.Request("configure", map[string]interface{}{
		"op":    "set",
		"path":  strings.Split(path, " "),
		"value": value,
	})
	return err
}

func (svc *ConfigService) SetTree(path string, value map[string]interface{}) error {
    _, err := svc.client.Request("configure", map[string]interface{}{
        "op": "set",
        "path": strings.Split(path, " "),
        "value": value,
    })
    return err
}

// Deletes the configuration tree/value at the specified path
func (svc *ConfigService) Delete(path string) error {
	_, err := svc.client.Request("configure", map[string]interface{}{
		"op":   "delete",
		"path": strings.Split(path, " "),
	})
	return err
}
