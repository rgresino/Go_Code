package dataManager

// dataManager package contains the functionality to record and retrieve the metrics
// including events & reasons of a service.

import (
	"encoding/json"

	"errors"

	"URL/executive-dashboard-dc/config"
	"UR/executive-dashboard-dc/logging"
	"URL/executive-dashboard-dc/server"
	"URL/executive-dashboard-dc/serviceMetrics"
)

// Used to update any insterested clients with entire dataset.
// TODO: we likely want to track the change sets and only send deltas
func BroadcastAllPings() {

	details, errors := FetchAllPings(true)
	if len(errors) != 0 {
		for _, err := range errors {
			logging.Error.Println(err.Error())
		}
	} else {
		broadcastPings(details)
	}
}

// used to broadcast provided pings to all interested clients
func broadcastPings(pings []serviceMetrics.ConfigurationPing) {
	b, err := json.Marshal(pings)
	if err != nil {
		logging.Error.Println("Error creating JSON representation of details data: " + err.Error())
	} else {
		// Broadcast to all interested clients
		server.BroadcastClientUpdate(string(b), server.UPDATE_CHANNEL)
	}
}

// Fetch a ping. If cache is to be used, the ping will come from the cache, but if not there, retrieve from graphite
func fetchPing(configurationId string, useCache bool) (result serviceMetrics.ConfigurationPing, err error) {
	ping, ok := retrieveCachedPing(configurationId)
	if !ok {
		return ping, errors.New("Configuration not found in Ping cache, configurationId: " + configurationId)
	}
	// Get the pager duty events for this service using the PagerDuty assigned service ID we have stored in our configuration
	// Tack on the pager duty events for this service
	ping.PagerDutyEvents = serviceMetrics.ReadOnePDService(configurationId)
	return ping, nil
}

// configurationIds identifies the configuration pings to return from the cache
// The error map identifies the cache misses
func FetchPings(configurationIds []string, useCache bool) ([]serviceMetrics.ConfigurationPing, map[string]error) {
	resultPings := []serviceMetrics.ConfigurationPing{}
	resultErrors := make(map[string]error)
	for _, configurationId := range configurationIds {
		pingFetched, err := fetchPing(configurationId, useCache)
		if err != nil {
			resultErrors[configurationId] = err
		} else {
			resultPings = append(resultPings, pingFetched)
		}
	}
	return resultPings, resultErrors
}

// FetchAllPings is a convenience method for retrieving the pings for all the configurations
// See FetchPings for more details
func FetchAllPings(useCache bool) ([]serviceMetrics.ConfigurationPing, map[string]error) {
	configurationIds := make([]string, 0, len(config.Configurations))
	for id := range config.Configurations {
		configurationIds = append(configurationIds, id)
	}
	return FetchPings(configurationIds, useCache)
}
