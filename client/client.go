package client

import (
	"encoding/json"
	"errors"
	"fmt"
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
func (c *Client) Request(endpoint string, payload interface{}) (interface{}, error) {
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

func _Flatten(data *map[string]string, path string, tree map[string]interface{}) error {
	for k, v := range tree {
		subpath := path
		if len(subpath) > 0 {
			subpath += " "
		}
		subpath += k

		subval, ok := v.(string)
		subtree, tok := v.(map[string]interface{})

		if ok {
			(*data)[subpath] = subval
		} else if tok {
			err := _Flatten(data, subpath, subtree)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s: Invalid type", subpath)
		}
	}

	return nil
}

// Flatten a multi level object into a flat list of keys and values
func Flatten(tree map[string]interface{}) (map[string]string, error) {
	res := map[string]string{}
	err := _Flatten(&res, "", tree)
	return res, err
}

// Sets all of the configuration keys and values in an object
func (svc *ConfigService) SetTree(tree map[string]interface{}) error {
	flat, err := Flatten(tree)
	if err != nil {
		return err
	}

	data := []map[string]interface{}{}
	for path, value := range flat {
		data = append(data, map[string]interface{}{
			"op":    "set",
			"path":  strings.Split(path, " "),
			"value": value,
		})
	}

	_, err = svc.client.Request("configure", data)
	return err
}

// Deletes the configuration tree/values at the specified paths
func (svc *ConfigService) Delete(paths ...string) error {
	data := []map[string]interface{}{}
	for _, path := range paths {
		data = append(data, map[string]interface{}{
			"op":   "delete",
			"path": strings.Split(path, " "),
		})
	}

	_, err := svc.client.Request("configure", data)
	return err
}
