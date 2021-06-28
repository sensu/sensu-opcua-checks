package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Endpoint  string
	Namespace string
	Format    string
	Tags      []string
	Nodes     []string
	DryRun    bool
	Verbose   bool
}

const (
	defaultMetricsFormat = "prometheus"
)

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-opcua-metrics",
			Short:    "Sensu metrics for OPC-UA nodes",
			Keyspace: "sensu.io/plugins/opcua-metrics/config",
		},
	}
	nodes = []*ua.ReadValueID{}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "endpoint",
			Env:       "OPCUA_ENDPOINT",
			Argument:  "endpoint",
			Shorthand: "e",
			Default:   "opc.tcp://localhost:4840",
			Usage:     "OP-CUA endpoint",
			Value:     &plugin.Endpoint,
		},
		&sensu.PluginConfigOption{
			Path:     "metrics-format",
			Env:      "OPCUA_METRICS_FORMAT",
			Argument: "metrics-format",
			Default:  defaultMetricsFormat,
			Usage:    "Metrics Format (currently only prometheus supported)",
			Value:    &plugin.Format,
		},
		&sensu.PluginConfigOption{
			Path:      "nodes",
			Env:       "OPCUA_NODES",
			Argument:  "nodes",
			Shorthand: "n",
			Default:   []string{},
			Usage:     "Comma separated list of OP-CUA nodes to read (ex \"ns=1;s=NODE_NAME1, ns=1;s=NODE_NAME2\")",
			Value:     &plugin.Nodes,
		},
		/*
			&sensu.PluginConfigOption{
				Path:      "namespace",
				Env:       "OPCUA_NAMESPACE",
				Argument:  "namespace",
				Shorthand: "N",
				Default:   "",
				Usage:     "The UA namespace, used with nodes to form full node name",
				Value:     &plugin.Namespace,
			},
		*/
		&sensu.PluginConfigOption{
			Path:      "tags",
			Env:       "OPCUA_TAGS",
			Argument:  "tags",
			Shorthand: "t",
			Default:   []string{},
			Usage:     "Comma separated list of additional metrics tags to add (ex: \"first=value, second=value\")",
			Value:     &plugin.Tags,
		},
		&sensu.PluginConfigOption{
			Path:     "dry-run",
			Argument: "dry-run",
			Default:  false,
			Usage:    "Dry-run, do not communicate with endpoint, implies verbose",
			Value:    &plugin.DryRun,
		},
		&sensu.PluginConfigOption{
			Path:      "verbose",
			Argument:  "verbose",
			Shorthand: "v",
			Default:   false,
			Usage:     "Enable verbose logging",
			Value:     &plugin.Verbose,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(plugin.Endpoint) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--endpoint or OPCUA_ENDPOINT environment variable is required")
	}
	if len(plugin.Nodes) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--nodes or OPCUA_NODES environment variable is required")
	}
	if plugin.Format != "prometheus" {
		return sensu.CheckStateWarning, fmt.Errorf("only prometheus metrics output format supported at this time")
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	//log.Println("executing check with --endpoint", plugin.Endpoint)
	//log.Println("executing check with --nodes", plugin.Nodes)
	resp, err := readNodes()
	if err != nil {
		return sensu.CheckStateCritical, nil
	}
	err = createMetrics(resp)
	if err != nil {
		return sensu.CheckStateCritical, nil
	}
	return sensu.CheckStateOK, nil
}

func createMetrics(resp *ua.ReadResponse) error {
	output := ""
	tags := ""
	for i, result := range resp.Results {
		if result.Status != ua.StatusOK {
			return fmt.Errorf("Result Status not OK: %v", result.Status)
		}
		tags = fmt.Sprintf("ns=%v", nodes[i].NodeID.Namespace())
		for _, t := range plugin.Tags {
			tags += fmt.Sprintf(" ,%s", strings.TrimSpace(t))
		}
		/* Auto detection of metric point timestamp precision using a heuristic with a 250-ish year cutoff */
		timestamp := result.SourceTimestamp.UnixNano()
		switch ts := math.Log10(float64(timestamp)); {
		case ts < 10:
			// assume timestamp is seconds convert to millisecond
			timestamp = time.Unix(timestamp, 0).UnixNano() / int64(time.Millisecond)
		case ts < 13:
			// assume timestamp is milliseconds
		case ts < 16:
			// assume timestamp is microseconds
			timestamp = (timestamp * 1000) / int64(time.Millisecond)
		default:
			// assume timestamp is nanoseconds
			timestamp = timestamp / int64(time.Millisecond)

		}
		nodeMetric := fmt.Sprintf("%s{%s} %v %v\n", nodes[i].NodeID.StringID(), tags, result.Value.Value(), timestamp)
		if plugin.Verbose {
			log.Printf(" %v\n", nodeMetric)
		}
		output += nodeMetric
	}
	fmt.Printf("%s\n", output)
	return nil
}

func readNodes() (*ua.ReadResponse, error) {
	var nodeID string
	if plugin.Verbose {
		log.Println("Reading Nodes:")
	}
	for _, n := range plugin.Nodes {
		nodeID = ""
		if len(plugin.Namespace) > 0 {
			nodeID = fmt.Sprintf("ns=%s;s=%s", plugin.Namespace, strings.TrimSpace(n))
		} else {
			nodeID = strings.TrimSpace(n)
		}
		if plugin.Verbose {
			log.Printf("  Node: %s\n", nodeID)
		}
		id, err := ua.ParseNodeID(nodeID)
		if err != nil {
			log.Println(`Error: "%s" has invalid node id: %v`, nodeID, err)
			return nil, err
		}
		nodes = append(nodes, &ua.ReadValueID{NodeID: id})
	}
	ctx := context.Background()
	c := opcua.NewClient(plugin.Endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err := c.Connect(ctx); err != nil {
		log.Println("Error: UA connection: %s", err)
		return nil, err
	}
	defer c.Close()

	req := &ua.ReadRequest{
		MaxAge:             2000,
		NodesToRead:        nodes,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.Read(req)
	if err != nil {
		log.Println("Error: UA Read Request failed: %s", err)
		return nil, err
	}
	return resp, err
}
