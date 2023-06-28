package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	Config          *ConfigService
	ContainerImages *ContainerImageService
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
		&sync.Mutex{},

		nil,
		nil,
	}

	client.Config = &ConfigService{client}
	client.ContainerImages = &ContainerImageService{client}

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

	url := c.url + "/" + endpoint
	c.mutex.Lock()
	resp, err := c.resty.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"key":  c.key,
			"data": string(data),
		}).
		Post(url)
	c.mutex.Unlock()
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode()
	body := resp.Body()
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf(
			"received non-successful (%d) response from vyos api (%s).\n%s",
			statusCode,
			url,
			body,
		)
	}

	r := new(response)
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	// Handle errors from the API
	if r.Error != nil {
		return nil, errors.New(*r.Error)
	}

	return r.Data, err
}

// Return the configuration tree at the specified path
func (svc *ConfigService) Show(ctx context.Context, path string) (any, error) {
	components := strings.Split(path, " ")
	terminal := components[len(components)-1]

	path_components := components
	if path == "" {
		path_components = []string{}
	}

	resp, err := svc.client.Request(ctx, "retrieve", map[string]any{
		"op":   "showConfig",
		"path": path_components,
	})
	if err != nil {
		if strings.Contains(err.Error(), "specified path is empty") {
			// If we get an empty path error, consume it and return nil
			return nil, nil
		} else {
			return nil, err
		}
	}

	obj, ok := resp.(map[string]any)
	if !ok {
		return nil, errors.New("received unexpected repsonse format from server")
	}

	val, ok := obj[terminal]
	if ok {
		return val, nil
	} else {
		return obj, nil
	}
}

// Set the configuration at the specified path.
//
// If `value` is a string it will be directly set. For lists maps, and any
// nesting of those types, each individual value will be set in a batch.
func (svc *ConfigService) Set(ctx context.Context, path string, value any) error {
	flat, err := Flatten(value)
	if err != nil {
		return err
	}

	payload := []map[string]any{}
	for _, pair := range flat {
		subpath, value := pair[0], pair[1]

		prefixpath := path
		if len(prefixpath) > 0 && len(subpath) > 0 {
			prefixpath += " "
		}
		prefixpath += subpath

		payload = append(payload, map[string]any{
			"op":    "set",
			"path":  strings.Split(prefixpath, " "),
			"value": value,
		})
	}

	_, err = svc.client.Request(ctx, "configure", payload)
	return err
}

// Delete values at the specified path.
//
// If `value` is nil or zero length the whole `path` will be deleted.
// If `value` is a string it will be directly deleted. For lists maps, and any
// nesting of those types, each individual value will be deleted in a batch.
func (svc *ConfigService) Delete(ctx context.Context, path string, value ...any) error {
	payload := []map[string]any{}

	if value == nil || len(value) == 0 {

		payload = append(payload, map[string]any{
			"op":   "delete",
			"path": strings.Split(path, " "),
		})
	} else {

		flat, err := Flatten(value)
		if err != nil {
			return err
		}

		for _, pair := range flat {
			subpath, value := pair[0], pair[1]

			prefixpath := path
			if len(prefixpath) > 0 && len(subpath) > 0 {
				prefixpath += " "
			}
			prefixpath += subpath

			payload = append(payload, map[string]any{
				"op":    "delete",
				"path":  strings.Split(prefixpath, " "),
				"value": value,
			})
		}
	}

	_, err := svc.client.Request(ctx, "configure", payload)
	return err
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
