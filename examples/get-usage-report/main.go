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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/laurentganne/yorc-provider-go-client/v1/yorcprovider"
	"github.com/pkg/errors"
)

// Command arguments
var url, user, password, orchestratorName, locationType, locationName string

type queryType struct {
	params map[string]string
}

var query queryType

func (q *queryType) String() string {
	return fmt.Sprintf("%+v", q.params)
}

func (q *queryType) Set(value string) error {
	v := strings.Split(value, "=")
	if len(v) != 2 {
		return errors.Errorf("Expected query parameter of the form key=value, got %s", value)
	}
	q.params[v[0]] = v[1]
	return nil
}

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&orchestratorName, "orchestrator", "", "Orchestrator name")
	flag.StringVar(&locationType, "type", "", "Location type")
	flag.StringVar(&locationName, "location", "", "Location")
	query.params = make(map[string]string)
	flag.Var(&query, "query", "Query parameter of the form \"key=value\" (you can use this flag mutiple times to define multiple query params)")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameters
	if orchestratorName == "" {
		log.Panic("Mandatory argument 'orchestrator' missing (Orchestrator name)")
	}
	if locationType == "" {
		log.Panic("Mandatory argument 'type' missing (Type of location for which to get a usage report)")
	}
	if locationName == "" {
		log.Panic("Mandatory argument 'location' missing (Name of location for which to get a usage report)")
	}

	client, err := yorcprovider.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	err = client.Login()
	if err != nil {
		log.Panic(err)
	}

	// Check the orchestrator specified exists
	orchestratorFound := false
	orchestrators, err := client.OrchestratorService().GetOrchestrators()
	if err != nil {
		log.Panic(err)
	}
	var orchestratorList []string
	for _, orchestrator := range orchestrators {
		orchestratorFound = (orchestrator.Name == orchestratorName)
		if orchestratorFound {
			break
		} else {
			orchestratorList = append(orchestratorList, orchestrator.Name)
		}
	}

	if !orchestratorFound {
		log.Panicf("No orchestrator %s found. Known orchestrators: %v", orchestratorName, orchestratorList)
	}

	// Get the collector for the expected location type
	var collectorID string
	collectors, err := client.UsageCollectorService().GetUsageCollectors(orchestratorName)
	if err != nil {
		log.Panic(err)
	}
	for _, collector := range collectors {
		if collector.ID == locationType {
			collectorID = collector.ID
			break
		}
	}
	if collectorID == "" {
		log.Panicf("Found no collector for %s on orchestrator %s", locationType, orchestratorName)
	}

	// Query a collection of resources usage
	queryID, err := client.UsageCollectorService().Query(orchestratorName, collectorID, locationName, query.params)
	if err != nil {
		log.Panic(err)
	}

	// Wait for the end of collection
	fmt.Printf("Waiting for the end of collection query...")
	done := false
	var collection *yorcprovider.UsageCollection
	for !done {
		time.Sleep(1 * time.Second)
		collection, err = client.UsageCollectorService().GetCollectedUsage(queryID)
		if err != nil {
			log.Panic(err)
		}

		done = (collection.Status == yorcprovider.QueryStatusDone ||
			collection.Status == yorcprovider.QueryStatusFailed ||
			collection.Status == yorcprovider.QueryStatusCanceled)
	}

	if collection.Status == yorcprovider.QueryStatusDone {
		fmt.Printf("\ncollection for %s location %s %s:\n%+s\n", orchestratorName, locationName, query.params, prettyPrint(collection.Results))
	} else {
		fmt.Printf("\nFailed to get collection for %s location %s %s: status %s\n", orchestratorName, locationName, query.params, collection.Status)
	}

	// Now that the query is done, deleting it
	err = client.UsageCollectorService().DeleteQuery(queryID)
	if err != nil {
		log.Panic(err)
	}

}

func prettyPrint(v interface{}) string {
	var result string
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		result = fmt.Sprintln(string(b))
	}
	return result
}
