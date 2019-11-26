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
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// UsageCollectorService is the interface to the service mamaging usage collectors
type UsageCollectorService interface {
	// Returns the list of usage collectors provided on a given orchestrator
	GetUsageCollectors(orchestratorName string) ([]UsageCollector, error)
	// Queries the collection of resources usage on a given location
	// The ID of a query that will perform the collection is returned
	Query(orchestratorName, collectorID, location string, queryParameters map[string]string) (string, error)
	// Deletes a query of resources usage collection
	DeleteQuery(queryID string) error
	// Gets queries of resources usage performed on a given orchestrator, for a given collector
	GetQueryIDs(orchestratorName, collectorID string) ([]string, error)
	// Gets results of a resources usage collection query
	GetCollectedUsage(queryID string) (*UsageCollection, error)
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
func (u *usageCollectorService) Query(orchestratorName, collectorID, location string, queryParameters map[string]string) (string, error) {

	var queryID string
	usageURL, err := url.Parse(fmt.Sprintf("%s/orchestrators/%s/infra_usage/%s/%s",
		yorcProviderRESTPrefix, orchestratorName, collectorID, location))
	if err != nil {
		return queryID, err
	}

	query := usageURL.Query()
	for k, v := range queryParameters {
		query.Set(k, v)
	}

	usageURL.RawQuery = query.Encode()

	response, err := u.client.do(
		"POST",
		usageURL.String(),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return queryID, errors.Wrapf(err, "Cannot send a request to submit a query on resources usage for %s %s %s",
			orchestratorName, collectorID, location)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return queryID, getError(response.Body)
	}

	locationHeader := response.Header["Location"]
	if len(locationHeader) == 0 || locationHeader[0] == "" {
		return queryID, errors.Wrapf(err, "No resources usage query could be created for %s %s %s",
			orchestratorName, collectorID, location)
	}

	queryIDPrefix := fmt.Sprintf("%s/orchestrators/", yorcProviderRESTPrefix)
	queryID = strings.TrimPrefix(locationHeader[0], queryIDPrefix)

	return queryID, err
}

// DeleteQuery deletes a query of resources usage collection
func (u *usageCollectorService) DeleteQuery(queryID string) error {
	response, err := u.client.do(
		"DELETE",
		fmt.Sprintf("%s/orchestrators/%s", yorcProviderRESTPrefix, queryID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to undeploy A4C application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// GetQueryIDs returns IDs of resources usage queries performed
// on a given orchestrator for a given collector
func (u *usageCollectorService) GetQueryIDs(orchestratorName, collectorID string) ([]string, error) {

	response, err := u.client.do(
		"GET",
		fmt.Sprintf("%s/orchestrators/%s/infra_usage", yorcProviderRESTPrefix, orchestratorName),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get query IDs on %s", orchestratorName)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read response to get query IDs on %s", orchestratorName)
	}

	var res struct {
		Data struct {
			Tasks []struct {
				Rel  string `json:"rel,omitempty"`
				HRef string `json:"href,omitempty"`
				Type string `json:"type,omitempty"`
			} `json:"tasks,omitempty"`
		} `json:"data"`
	}
	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of response to get query IDs on %s", orchestratorName)
	}

	// Getting query IDs from href
	var result []string
	queryIDPrefix := fmt.Sprintf("%s/orchestrators/", yorcProviderRESTPrefix)
	for _, t := range res.Data.Tasks {
		s := strings.TrimPrefix(t.HRef, queryIDPrefix)
		if collectorID != "" {
			// String format <orchestrator>/infra_usage/<collector>/tasks/<id>
			values := strings.Split(s, "/")
			if len(values) > 3 || values[2] != collectorID {
				// This query is for another collector
				break
			}
		}
		result = append(result, s)
	}
	return result, err
}

// GetCollectedUsage gets results of a resources usage collection query
func (u *usageCollectorService) GetCollectedUsage(queryID string) (*UsageCollection, error) {
	response, err := u.client.do(
		"GET",
		fmt.Sprintf("%s/orchestrators/%s", yorcProviderRESTPrefix, queryID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get usage collected by query %s", queryID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read response to get usage collected by query %s", queryID)
	}

	var res struct {
		Data struct {
			ID       string                 `json:"id,omitempty"`
			TargetID string                 `json:"target_id,omitempty"`
			Type     string                 `json:"type,omitempty"`
			Status   string                 `json:"status,omitempty"`
			Results  map[string]interface{} `json:"result_set,omitempty"`
		} `json:"data"`
	}
	if err = json.Unmarshal(responseBody, &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of response to get collectors on %s: %s", queryID, string(responseBody))
	}

	result := UsageCollection{
		Status:  res.Data.Status,
		Results: res.Data.Results,
	}
	return &result, err
}
