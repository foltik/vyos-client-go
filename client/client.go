package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	url   string
	key   string
	resty *resty.Client

	mutex *sync.Mutex

	Config *ConfigService
}
type ConfigService struct {
	client        *Client
	mutex         *sync.Mutex
	showTreeCache *map[string]any
}

func New(url string, key string) *Client {
	return NewWithClient(&http.Client{Timeout: 10 * time.Second}, url, key)
}

func NewWithClient(c *http.Client, url string, key string) *Client {
	client := &Client{
		url,
		key,
		resty.NewWithClient(c),
		&sync.Mutex{},

		nil,
	}

	client.Config = &ConfigService{
		client,
		&sync.Mutex{},
		nil,
	}

	return client
}

type response struct {
	Success bool
	Data    any
	Error   *string
}

// Post a raw request with `payload` to `endpoint`.
func (c *Client) Request(ctx context.Context, endpoint string, payload any) (any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("failed to marshal request payload")
	}

	c.mutex.Lock()
	resp, err := c.resty.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"key":  c.key,
			"data": string(data),
		}).
		Post(c.url + "/" + endpoint)
	c.mutex.Unlock()
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

// Return the full configuration tree at the specified path
func (svc *ConfigService) ShowAny(ctx context.Context, path string) (any, error) {
	svc.mutex.Lock()
	if svc.showTreeCache == nil {
		resp, err := svc.client.Request(ctx, "retrieve", map[string]any{
			"op":   "showConfig",
			"path": []string{},
		})
		if err != nil {
			svc.mutex.Unlock()
			if strings.Contains(err.Error(), "specified path is empty") {
				// If we get an empty path error, consume it and return nil
				return nil, nil
			} else {
				return nil, err
			}
		}

		obj, ok := resp.(map[string]any)
		if !ok {
			svc.mutex.Unlock()
			return nil, errors.New("received unexpected repsonse format from server")
		}
		svc.showTreeCache = &obj
	}
	svc.mutex.Unlock()

	var val any = *svc.showTreeCache
	for _, component := range strings.Split(path, " ") {
		obj, ok := val.(map[string]any)[component]
		if !ok {
			return nil, nil
		}
		val = obj
	}

	return val, nil
}

func (svc *ConfigService) ShowTree(ctx context.Context, path string) (map[string]any, error) {

	obj, err := svc.ShowAny(ctx, path)
	if err != nil {
		return nil, err
	}

	result, ok := obj.(map[string]any)
	if !ok {
		return nil, errors.New("Path '" + path + "' exists but it is not a tree")
	}

	return result, nil
}

// Return the single configuration value at the speicfied path
func (svc *ConfigService) Show(ctx context.Context, path string) (*string, error) {

	obj, err := svc.ShowAny(ctx, path)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}

	val, ok := obj.(string)
	if !ok {
		return nil, errors.New("value missing from configuration tree returned by server")
	}

	return &val, nil
}

// Save the running configuration to the default startup configuration
func (svc *ConfigService) Save(ctx context.Context) error {
	_, err := svc.client.Request(ctx, "config-file", map[string]any{
		"op": "save",
	})
	return err
}

// Save the running configuration to the specified file
func (svc *ConfigService) SaveFile(ctx context.Context, file string) error {
	_, err := svc.client.Request(ctx, "config-file", map[string]any{
		"op":   "save",
		"file": file,
	})
	return err
}

// Load a configuration file
func (svc *ConfigService) LoadFile(ctx context.Context, file string) error {
	_, err := svc.client.Request(ctx, "config-file", map[string]any{
		"op":   "load",
		"file": file,
	})
	return err
}

// Set a configuration value at the specified path
func (svc *ConfigService) Set(ctx context.Context, path string, value string) error {
	_, err := svc.client.Request(ctx, "configure", map[string]any{
		"op":    "set",
		"path":  strings.Split(path, " "),
		"value": value,
	})
	return err
}

// Delete the configuration tree/values at the specified paths
func (svc *ConfigService) Delete(ctx context.Context, paths ...string) error {
	payload := []map[string]any{}
	for _, path := range paths {
		payload = append(payload, map[string]any{
			"op":   "delete",
			"path": strings.Split(path, " "),
		})
	}

	_, err := svc.client.Request(ctx, "configure", payload)
	return err
}

// Set all of the configuration keys and values in an object
func (svc *ConfigService) SetTree(ctx context.Context, tree map[string]any) error {
	flat, err := Flatten(tree)
	if err != nil {
		return err
	}

	return svc.SetFlat(ctx, flat)
}

// Set all of the configuration in a flat list of {key, value} pairs
func (svc *ConfigService) SetFlat(ctx context.Context, flat [][]string) error {
	payload := []map[string]any{}
	for _, pair := range flat {
		path, value := pair[0], pair[1]
		payload = append(payload, map[string]any{
			"op":    "set",
			"path":  strings.Split(path, " "),
			"value": value,
		})
	}

	_, err := svc.client.Request(ctx, "configure", payload)
	return err
}

// Delete all of the configuration keys in an object
func (svc *ConfigService) DeleteTree(ctx context.Context, tree map[string]any) error {
	flat, err := Flatten(tree)
	if err != nil {
		return err
	}

	return svc.DeleteFlat(ctx, flat)
}

// Delete all of the configuration in a flat list of {key, value} pairs
func (svc *ConfigService) DeleteFlat(ctx context.Context, flat [][]string) error {
	payload := []map[string]any{}
	for _, pair := range flat {
		path, value := pair[0], pair[1]

		target := path
		if value != "" {
			target += " " + value
		}

		payload = append(payload, map[string]any{
			"op":   "delete",
			"path": strings.Split(target, " "),
		})
	}

	_, err := svc.client.Request(ctx, "configure", payload)
	return err
}
