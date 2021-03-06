/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/hashicorp/nomad, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package command

import (
	"fmt"
	"strings"
)

type ConfigCommand struct {
	Meta
}

func (c *ConfigCommand) Help() string {
	helpText := `
Usage: dtle agent-config [options]

  View or modify client configuration details. This command only
  works on client nodes, and can be used to update the running
  client configurations it supports.

  The arguments behave differently depending on the flags given.
  See each flag's description for its specific requirements.

General Options:

  ` + generalOptionsUsage() + `

Agent Config Options:

  -servers
    List the known server addresses of the client node. Client
    nodes do not participate in the gossip pool, and instead
    register with these servers periodically over the network.

  -update-servers
    Updates the agent's server list using the provided
    arguments. Multiple server addresses may be passed using
    multiple arguments. IMPORTANT: When updating the servers
    list, you must specify ALL of the server nodes you wish
    to configure. The set is updated atomically.

    Example:
      $ dtle agent-config -update-servers foo:8191 bar:8191
`
	return strings.TrimSpace(helpText)
}

func (c *ConfigCommand) Synopsis() string {
	return "View or modify agent configuration details"
}

func (c *ConfigCommand) Run(args []string) int {
	var listServers, updateServers bool

	flags := c.Meta.FlagSet("client-servers", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.BoolVar(&listServers, "servers", false, "")
	flags.BoolVar(&updateServers, "update-servers", false, "")

	if err := flags.Parse(args); err != nil {
		return 1
	}
	args = flags.Args()

	// Check the flags for misuse
	if !listServers && !updateServers {
		c.Ui.Error(c.Help())
		return 1
	}

	// Get the HTTP client
	client, err := c.Meta.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	if updateServers {
		// Get the server addresses
		if len(args) == 0 {
			c.Ui.Error(c.Help())
			return 1
		}

		// Set the servers list
		if err := client.Agent().SetServers(args); err != nil {
			c.Ui.Error(fmt.Sprintf("Error updating server list: %s", err))
			return 1
		}
		c.Ui.Output(fmt.Sprint("Updated server list"))
		return 0
	}

	if listServers {
		// Query the current server list
		servers, err := client.Agent().Servers()
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error querying server list: %s", err))
			return 1
		}

		// Print the results
		for _, server := range servers {
			c.Ui.Output(server)
		}
		return 0
	}

	// Should not make it this far
	return 1
}
