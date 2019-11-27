// Copyright 2019 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yorcprovider

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
	"github.com/pkg/errors"
)

// Client is the client interface to the Yorc Provider
type Client interface {
	Login() error
	Logout() error
	OrchestratorService() OrchestratorService
	UsageCollectorService() UsageCollectorService
}

const (
	// QueryStatusInitial is the initial status of a qurery
	QueryStatusInitial = "INITIAL"
	// QueryStatusRunning is the status of query running (in the process of collecting usage)
	QueryStatusRunning = "RUNNING"
	// QueryStatusDone is the status of a query for which the work of data collection is done
	QueryStatusDone = "DONE"
	// QueryStatusFailed is the status of a query for which the work of data collection failed
	QueryStatusFailed = "FAILED"
	// QueryStatusCanceled is the status of a query for which the work of data collection was canceled
	QueryStatusCanceled = "CANCELED"
)

const (
	yorcProviderRESTPrefix = "/rest/yorc-collector-plugin/latest"
)

// NewClient instanciates and returns Client
func NewClient(a4cURL string, user string, password string, caFile string, skipSecure bool) (Client, error) {
	a4cAPI := strings.TrimRight(a4cURL, "/")

	if m, _ := regexp.Match("^http[s]?://.*", []byte(a4cAPI)); !m {
		a4cAPI = "http://" + a4cAPI
	}

	var useTLS = true
	if m, _ := regexp.Match("^http://.*", []byte(a4cAPI)); m {
		useTLS = false
	}

	url, err := urlx.Parse(a4cAPI)
	if err != nil {
		return nil, errors.Wrapf(err, "Malformed alien4cloud URL: %s", a4cAPI)
	}

	a4chost, _, err := urlx.SplitHostPort(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Malformed alien4cloud URL %s", url)
	}

	tlsConfig := &tls.Config{ServerName: a4chost}

	if useTLS {
		if caFile == "" || skipSecure {
			if skipSecure {
				tlsConfig.InsecureSkipVerify = true
			} else {
				return nil, errors.Errorf("You must provide a certificate authority file in TLS verify mode")
			}
		}

		if !skipSecure {
			certPool := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to read certificate authority file")
			}
			if !certPool.AppendCertsFromPEM(caCert) {
				return nil, errors.Errorf("%q is not a valid certificate authority.", caCert)
			}
			tlsConfig.RootCAs = certPool
		}
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}

	restClient := restClient{
		Client: &http.Client{
			Transport:     tr,
			CheckRedirect: nil,
			Jar:           newJar(),
			Timeout:       0},
		baseURL:  a4cAPI,
		username: user,
		password: password,
	}
	return &yorcProviderClient{
		client:                restClient,
		orchestratorService:   &orchestratorService{restClient},
		usageCollectorService: &usageCollectorService{restClient},
	}, nil
}

// Login login to alien4cloud
func (c *yorcProviderClient) Login() error {
	return c.client.login()
}

// Logout log out from alien4cloud
func (c *yorcProviderClient) Logout() error {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/logout", c.client.baseURL), nil)
	if err != nil {
		log.Panic(err)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Set("Connection", "close")

	request.Close = true

	response, err := c.client.Client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// OrchestratorService retrieves the Orchestrator Service
func (c *yorcProviderClient) OrchestratorService() OrchestratorService {
	return c.orchestratorService
}

// UsageCollectorService retrieves the Orchestrator Service
func (c *yorcProviderClient) UsageCollectorService() UsageCollectorService {
	return c.usageCollectorService
}

type restClient struct {
	*http.Client
	baseURL  string
	username string
	password string
}

type yorcProviderClient struct {
	client                restClient
	orchestratorService   *orchestratorService
	usageCollectorService *usageCollectorService
}

// do requests the alien4cloud rest api with a Context that can be canceled
func (r *restClient) doWithContext(ctx context.Context, method string, path string, body []byte, headers []Header) (*http.Response, error) {

	bodyBytes := bytes.NewBuffer(body)

	// Create the request
	var request *http.Request
	var err error
	if ctx == nil {
		request, err = http.NewRequest(method, r.baseURL+path, bodyBytes)
	} else {

		request, err = http.NewRequestWithContext(ctx, method, r.baseURL+path, bodyBytes)
	}

	if err != nil {
		return nil, err
	}

	// Add header
	for _, header := range headers {
		request.Header.Add(header.Key, header.Value)
	}

	response, err := r.Client.Do(request)
	if err != nil {
		return nil, err
	}

	// Cookie can potentially be expired. If we are unauthorized to send a request, we should try to login again.
	if response.StatusCode == http.StatusForbidden {
		err = r.login()
		if err != nil {
			return nil, err
		}

		bodyBytes = bytes.NewBuffer(body)

		request, err := http.NewRequest(method, r.baseURL+path, bodyBytes)
		if err != nil {
			return nil, err
		}

		for _, header := range headers {
			request.Header.Add(header.Key, header.Value)
		}

		response, err := r.Client.Do(request)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	return response, nil
}

// do requests the alien4cloud rest api
func (r *restClient) do(method string, path string, body []byte, headers []Header) (*http.Response, error) {

	return r.doWithContext(nil, method, path, body, headers)
}

// login to alien4cloud
func (r *restClient) login() error {
	values := url.Values{}
	values.Set("username", r.username)
	values.Set("password", r.password)
	values.Set("submit", "Login")
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/login", r.baseURL),
		strings.NewReader(values.Encode()))
	if err != nil {
		log.Panic(err)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := r.Client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}
