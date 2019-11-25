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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// UsageCollectorService is the interface to the service mamaging usage collectors
type UsageCollectorService interface {
	// Returns the list of usage collectors provided on a given orchestrator
	GetUsageCollectors(orchestratorName string) ([]UsageCollector, error)
	// Queries the collection of resources usage on a given location
	// The ID of a query that will perform the collection is returned
	Query(collectorID string, location string, queryParameters map[string]string) (string, error)
	// Deletes a query of resources usage collection
	DeleteQuery(collectorID string, location string, queryID string) error
	// Gets queries of resources usahe performed on a given collector
	GetQueries(collectorID string) (map[string][]string, error)
	// Gets results of a resources usage collection query
	GetQueryCollectedUsage(collectorID string, location string, queryID string) (map[string]interface{}, error)
}

type usageCollectorService struct {
	client restClient
}

// GetUsageCollectors returns the list of usage collectors provided on a given orchestrator
func (u *usageCollectorService) GetUsageCollectors(orchestratorName string) ([]UsageCollector, error) {

	// Get orchestrator location
	response, err := u.client.do(
		"GET",
		fmt.Sprintf("%s/orchestrators/%s/registry/infra_usage_collectors", yorcProviderRESTPrefix, orchestratorName),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get collectors on %s", orchestratorName)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read response to get collectors on %s", orchestratorName)
	}

	var res struct {
		Data struct {
			Infrastructures []UsageCollector `json:"infrastructures,omitempty"`
		} `json:"data"`
	}
	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of response to get collectors on %s", orchestratorName)
	}

	return res.Data.Infrastructures, err
}

// Queries the collection of resources usage on a given location
// The ID of a query that will perform the collection is returned
func (u *usageCollectorService) Query(collectorID string, location string, queryParameters map[string]string) (string, error) {
	var err error
	return "", err
}

// DeleteQuery deletes a query of resources usage collection
func (u *usageCollectorService) DeleteQuery(collectorID string, location string, queryID string) error {
	var err error
	return err
}

// GetQueries gets queries of resources usahe performed on a given collector
func (u *usageCollectorService) GetQueries(collectorID string) (map[string][]string, error) {
	var err error
	return nil, err
}

// GetQueryCollectedUsage gets results of a resources usage collection query
func (u *usageCollectorService) GetQueryCollectedUsage(collectorID string, location string, queryID string) (map[string]interface{}, error) {
	var err error
	return nil, err
}
