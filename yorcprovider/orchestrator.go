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

// OrchestratorService is the interface to the service mamaging orchestrators
type OrchestratorService interface {
	// Returns the list of Yorc orchestrators configured
	GetOrchestrators() ([]Orchestrator, error)
}

type orchestratorService struct {
	client restClient
}

// GetOrchestrators returns the list of Yorc orchestrators configured
func (o *orchestratorService) GetOrchestrators() ([]Orchestrator, error) {

	// Get orchestrator location
	response, err := o.client.do(
		"GET",
		fmt.Sprintf("%s/orchestrators", yorcProviderRESTPrefix),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get orchestrators")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read response to get the list of orchestrators")
	}

	var res struct {
		Data struct {
			Orchestrators []Orchestrator `json:"orchestrators,omitempty"`
		} `json:"data"`
	}
	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of response to get the list of orchestrators")
	}

	return res.Data.Orchestrators, err
}
