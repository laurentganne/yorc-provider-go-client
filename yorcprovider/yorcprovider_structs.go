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

// Orchestrator holds properties describing an orchestrator
type Orchestrator struct {
	Name string `json:"name,omitempty"`
	HRef string `json:"href,omitempty"`
}

// UsageCollector holds properties describing a Usage Collector: its id, and the plugin
// implementing this collector
type UsageCollector struct {
	ID     string `json:"id,omitempty"`
	Origin string `json:"origin,omitempty"`
}

// DataCollection holds the status of a Resources usage query, and results when the
// collection is done
type DataCollection struct {
	Status  string                 `json:"status,omitempty"`
	Results map[string]interface{} `json:"results,omitempty"`
}

// Header is the representation of an http header
type Header struct {
	Key   string
	Value string
}

// Error is the representation of a yorc provider error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
