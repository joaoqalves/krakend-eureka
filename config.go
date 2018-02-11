package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/devopsfaith/krakend/config"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "github_com/joaoqalves/krakend-eureka"

var (
	// ErrNoConfig is the error to be returned when there is no config with the etcd namespace
	ErrNoConfig = fmt.Errorf("unable to create the etcd client: no config")
	// ErrBadConfig is the error to be returned when the config is not well defined
	ErrBadConfig = fmt.Errorf("unable to create the etcd client with the received config")
	// ErrNoMachines is the error to be returned when the config has not defined one or more servers
	ErrNoMachines = fmt.Errorf("unable to create the etcd client without a set of servers")
)

// New creates an etcd client with the config extracted from the extra config param
func New(ctx context.Context, e config.ExtraConfig) (Client, error) {
	v, ok := e[Namespace]
	if !ok {
		return nil, ErrNoConfig
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return nil, ErrBadConfig
	}
	machines, err := parseMachines(tmp)
	if err != nil {
		return nil, err
	}

	return NewClient(ctx, machines, parseOptions(tmp))
}

func parseMachines(cfg map[string]interface{}) ([]string, error) {
	result := []string{}
	machines, ok := cfg["machines"]
	if !ok {
		return result, ErrNoMachines
	}
	ms, ok := machines.([]interface{})
	if !ok {
		return result, ErrNoMachines
	}
	for _, m := range ms {
		if machine, ok := m.(string); ok {
			result = append(result, machine)
		}
	}
	if len(result) == 0 {
		return result, ErrNoMachines
	}
	return result, nil
}

func parseOptions(cfg map[string]interface{}) ClientOptions {
	options := ClientOptions{}
	v, ok := cfg["options"]
	if !ok {
		return options
	}
	tmp := v.(map[string]interface{})

	if o, ok := tmp["dial_timeout"]; ok {
		if d, err := parseDuration(o); err == nil {
			options.DialTimeout = d
		}
	}
	return options
}

func parseDuration(v interface{}) (time.Duration, error) {
	s, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("unable to parse %v as a time.Duration\n", v)
	}
	return time.ParseDuration(s)
}