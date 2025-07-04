package coredns_mikrotik_dhcp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Lease struct {
	Address  net.IP `json:"address"`
	Hostname string `json:"host-name"`
}

type LeaseGetter interface {
	GetBoundLeases(ctx context.Context) ([]Lease, error)
}

type MikroTikAPILeaseGetter struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type MikroTikAPILeaseGetterOption func(g *MikroTikAPILeaseGetter)

func WithInsecureSkipVerify() MikroTikAPILeaseGetterOption {
	return func(g *MikroTikAPILeaseGetter) {
		g.client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
}

func NewMikroTikAPILeaseGetter(baseURL string, username, password string, opts ...MikroTikAPILeaseGetterOption) *MikroTikAPILeaseGetter {
	g := &MikroTikAPILeaseGetter{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func (g *MikroTikAPILeaseGetter) doRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.SetBasicAuth(g.username, g.password)

	res, err := g.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errResponse errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
			return fmt.Errorf("unknown error, status code [%d]: %w", res.StatusCode, err)
		}

		message := "error occurred while calling MikroTik API: " + errResponse.Message
		if errResponse.Detail != "" {
			message += ": " + errResponse.Detail
		}

		return errors.New(message)
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (g *MikroTikAPILeaseGetter) GetBoundLeases(ctx context.Context) ([]Lease, error) {
	rawURL, err := url.JoinPath(g.baseURL, "/rest/ip/dhcp-server/lease")
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("status", "bound")
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	var res []Lease
	if err := g.doRequest(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type errorResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}
