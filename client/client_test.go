package client

import (
	"fmt"
	"testing"

	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
)

func make_client(t *testing.T) (*Client, context.Context) {
	// Create a CA pool which trusts the VyOS image's selfsigned cert
	pool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("../.github/workflows/selfsigned.pem")
	if err != nil {
		t.Errorf("failed to read selfsigned cert file: %s", err.Error())
	}
	pool.AppendCertsFromPEM(pem)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
	http := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	client := NewWithClient(http, "https://localhost", "vyos")
	ctx := context.Background()

	return client, ctx
}

func TestUnit_NewClient(t *testing.T) {
	make_client(t)
}

func TestIntegration_Config_ShowValue(t *testing.T) {
	client, ctx := make_client(t)
	resp, err := client.Config.Show(ctx, "service https virtual-host vyos server-name")
	if err != nil {
		t.Error(err)
	}

	server_name, ok := resp.(string)
	if !ok {
		t.Errorf("expected string, got '%s'", resp)
	}

	if server_name != "vyos.local" {
		t.Errorf("expected 'vyos.local', got '%s'", server_name)
	}
}

func TestIntegration_Config_ShowMap(t *testing.T) {
	client, ctx := make_client(t)
	resp, err := client.Config.Show(ctx, "service https api keys id apikey")
	if err != nil {
		t.Error(err)
	}

	config, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T: '%s'", resp, resp)
	}

	value, ok := config["key"]
	if !ok {
		t.Errorf("expected config to have key 'system', got %s", config)
	}
	if value != "vyos" {
		t.Errorf("expected 'vyos', got '%s'", value)
	}
}

func TestIntegration_Config_ShowArray(t *testing.T) {
	client, ctx := make_client(t)
	resp, err := client.Config.Show(ctx, "system name-server")
	if err != nil {
		t.Error(err)
	}

	nameservers, ok := resp.([]any)
	if !ok {
		t.Errorf("expected []any, got %T: '%s'", resp, resp)
	}

	if len(nameservers) < 2 {
		t.Errorf("expected array of length >=2, got %s", nameservers)
	}
	if nameservers[0] != "1.1.1.1" || nameservers[1] != "1.0.0.1" {
		t.Errorf("expected [1.1.1.1 1.0.0.1], got %s", nameservers)
	}
}

// Show("") should return the entire root config
func TestIntegration_Config_ShowRoot(t *testing.T) {
	client, ctx := make_client(t)
	resp, err := client.Config.Show(ctx, "")
	if err != nil {
		t.Error(err)
	}

	config, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T: '%s'", resp, resp)
	}

	_, ok = config["system"]
	if !ok {
		t.Errorf("expected config to have key 'system', got %s", config)
	}
}

func TestIntegration_Config_SetValue(t *testing.T) {
	client, ctx := make_client(t)
	err := client.Config.Set(ctx, "service ntp listen-address", "1.2.3.4")
	if err != nil {
		t.Error(err)
	}

	resp, err := client.Config.Show(ctx, "service ntp listen-address")
	if err != nil {
		t.Error(err)
	}

	hostname, ok := resp.(string)
	if !ok {
		t.Errorf("expected string, got '%s'", resp)
	}

	if hostname != "1.2.3.4" {
		t.Errorf("expected '1.2.3.4', got '%s'", hostname)
	}

	err = client.Config.Delete(ctx, "service ntp listen-address")
	if err != nil {
		t.Error(err)
	}
}

func TestIntegration_Config_SetMap(t *testing.T) {
	client, ctx := make_client(t)

	payload := map[string]any{
		"reboot-on-panic": "",
		"startup-beep":    "",
	}
	err := client.Config.Set(ctx, "system option", payload)
	if err != nil {
		t.Error(err)
	}

	resp, err := client.Config.Show(ctx, "system option")
	if err != nil {
		t.Error(err)
	}

	options, ok := resp.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T: '%s'", resp, resp)
	}

	_, ok = options["reboot-on-panic"]
	if !ok {
		t.Errorf("expected config to have key 'reboot-on-panic', got %s", options)
	}

	_, ok = options["startup-beep"]
	if !ok {
		t.Errorf("expected config to have key 'startup-beep', got %s", options)
	}

	err = client.Config.Delete(ctx, "system option")
	if err != nil {
		t.Error(err)
	}
}

func TestIntegration_Config_SetArray(t *testing.T) {
	client, ctx := make_client(t)

	payload := []string{"vyos.io", "vyos.net"}
	err := client.Config.Set(ctx, "system domain-search domain", payload)
	if err != nil {
		t.Error(err)
	}

	resp, err := client.Config.Show(ctx, "system domain-search domain")
	if err != nil {
		t.Error(err)
	}

	domains, ok := resp.([]any)
	if !ok {
		t.Errorf("expected []any, got %T: '%s'", resp, resp)
	}

	if len(domains) != 2 {
		t.Errorf("expected array of length 2, got %s", domains)
	}
	if domains[0] != "vyos.io" || domains[1] != "vyos.net" {
		t.Errorf("expected [vyos.io vyos.net], got %s", domains)
	}

	err = client.Config.Delete(ctx, "system domain-search")
	if err != nil {
		t.Error(err)
	}
}

// Set("", map) should merge the map into the root config
func TestIntegration_Config_SetRoot(t *testing.T) {
	client, ctx := make_client(t)

	payload := map[string]any{
		"system": map[string]any{
			"domain-name": "vyos.local",
		},
	}
	err := client.Config.Set(ctx, "", payload)
	if err != nil {
		t.Error(err)
	}

	resp, err := client.Config.Show(ctx, "service https virtual-host vyos server-name")
	if err != nil {
		t.Error(err)
	}

	domain, ok := resp.(string)
	if !ok {
		t.Errorf("expected string, got '%s'", resp)
	}

	if domain != "vyos.local" {
		t.Errorf("expected 'vyos.local', got '%s'", domain)
	}

	err = client.Config.Delete(ctx, "system domain-name")
	if err != nil {
		t.Error(err)
	}
}

func TestIntegration_Config_Delete(t *testing.T) {
	client, ctx := make_client(t)

	err := client.Config.Delete(ctx, "system host-name")
	if err != nil {
		t.Error(err)
	}

	hostname, err := client.Config.Show(ctx, "system host-name")
	if err != nil {
		t.Error(err)
	}

	if hostname != nil {
		t.Errorf("expected nil, got '%s'", hostname)
	}

	err = client.Config.Set(ctx, "system host-name", "vyos")
	if err != nil {
		t.Error(err)
	}
}

func TestIntegration_HandleNonSuccess(t *testing.T) {
	client, ctx := make_client(t)

	_, err := client.Request(ctx, "foo", map[string]any{
		"op": "foo",
	})
	assert.ErrorContains(
		t,
		err,
		fmt.Sprintf(
			"received non-successful (404) response from vyos api (%s/foo)",
			client.url,
		),
	)
}
