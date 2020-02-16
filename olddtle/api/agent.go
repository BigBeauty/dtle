/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/actiontech/dtle, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package api

import (
	"fmt"
	"net/url"
)

// Agent encapsulates an API client which talks to Udup's
// agent endpoints for a specific node.
type Agent struct {
	client *Client

	// Cache static agent info
	nodeName   string
	datacenter string
	region     string
}

// AgentCheck represents a check known to the agent
type AgentCheck struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
}

// AgentService represents a service known to the agent
type AgentService struct {
	ID                string
	Service           string
	Tags              []string
	Port              int
	Address           string
	EnableTagOverride bool
}

// Agent returns a new agent which can be used to query
// the agent-specific endpoints.
func (c *Client) Agent() *Agent {
	return &Agent{client: c}
}

// Self is used to query the /v1/self endpoint and
// returns information specific to the running agent.
func (a *Agent) Self() (*AgentSelf, error) {
	var out *AgentSelf

	// Query the self endpoint on the agent
	_, err := a.client.query("/v1/self", &out, nil)
	if err != nil {
		return nil, fmt.Errorf("failed querying self endpoint: %s", err)
	}

	// Populate the cache for faster queries
	a.populateCache(out)

	return out, nil
}

// populateCache is used to insert various pieces of static
// data into the agent handle. This is used during subsequent
// lookups for the same data later on to save the round trip.
func (a *Agent) populateCache(self *AgentSelf) {
	if a.nodeName == "" {
		a.nodeName = self.Member.Name
	}
	if a.datacenter == "" {
		if val, ok := self.Config["Datacenter"]; ok {
			a.datacenter, _ = val.(string)
		}
	}
	if a.region == "" {
		if val, ok := self.Config["Region"]; ok {
			a.region, _ = val.(string)
		}
	}
}

// NodeName is used to query the Udup agent for its node name.
func (a *Agent) NodeName() (string, error) {
	// Return from cache if we have it
	if a.nodeName != "" {
		return a.nodeName, nil
	}

	// Query the node name
	_, err := a.Self()
	return a.nodeName, err
}

// Datacenter is used to return the name of the datacenter which
// the agent is a member of.
func (a *Agent) Datacenter() (string, error) {
	// Return from cache if we have it
	if a.datacenter != "" {
		return a.datacenter, nil
	}

	// Query the agent for the DC
	_, err := a.Self()
	return a.datacenter, err
}

// Region is used to look up the region the agent is in.
func (a *Agent) Region() (string, error) {
	// Return from cache if we have it
	if a.region != "" {
		return a.region, nil
	}

	// Query the agent for the region
	_, err := a.Self()
	return a.region, err
}

// Members is used to query all of the known server members
func (a *Agent) Members() (*ServerMembers, error) {
	var resp *ServerMembers

	// Query the known members
	_, err := a.client.query("/v1/members", &resp, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ForceLeave is used to eject an existing node from the cluster.
func (a *Agent) ForceLeave(node string) error {
	_, err := a.client.write("/v1/agent/force-leave?node="+node, nil, nil, nil)
	return err
}

// Servers is used to query the list of servers on a client node.
func (a *Agent) Servers() ([]string, error) {
	var resp []string
	_, err := a.client.query("/v1/servers", &resp, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SetServers is used to update the list of servers on a client node.
func (a *Agent) SetServers(addrs []string) error {
	// Accumulate the addresses
	v := url.Values{}
	for _, addr := range addrs {
		v.Add("address", addr)
	}

	_, err := a.client.write("/v1/servers?"+v.Encode(), nil, nil, nil)
	return err
}

// joinResponse is used to decode the response we get while
// sending a member join request.
type joinResponse struct {
	NumJoined int    `json:"num_joined"`
	Error     string `json:"error"`
}

type ServerMembers struct {
	ServerName   string
	ServerRegion string
	ServerDC     string
	Members      []*AgentMember
}

type AgentSelf struct {
	Config map[string]interface{}       `json:"config"`
	Member AgentMember                  `json:"member"`
	Stats  map[string]map[string]string `json:"stats"`
}

// AgentMember represents a cluster member known to the agent
type AgentMember struct {
	Name        string
	Addr        string
	Port        uint16
	Tags        map[string]string
	Status      string
	ProtocolMin uint8
	ProtocolMax uint8
	ProtocolCur uint8
	DelegateMin uint8
	DelegateMax uint8
	DelegateCur uint8
}

// AgentMembersNameSort implements sort.Interface for []*AgentMembersNameSort
// based on the Name, DC and Region
type AgentMembersNameSort []*AgentMember

func (a AgentMembersNameSort) Len() int      { return len(a) }
func (a AgentMembersNameSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AgentMembersNameSort) Less(i, j int) bool {
	if a[i].Tags["region"] != a[j].Tags["region"] {
		return a[i].Tags["region"] < a[j].Tags["region"]
	}

	if a[i].Tags["dc"] != a[j].Tags["dc"] {
		return a[i].Tags["dc"] < a[j].Tags["dc"]
	}

	return a[i].Name < a[j].Name

}
