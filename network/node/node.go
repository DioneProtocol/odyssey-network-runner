package node

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanche-network-runner/api"
	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/ids"
)

type Config struct {
	// Configuration specific to a particular implementation of a node.
	ImplSpecificConfig interface{}
	// A node's name must be unique from all other nodes
	// in a network. If Name is the empty string, a
	// unique name is assigned on node creation.
	Name string `json:"name"`
	// True if other nodes should use this node
	// as a bootstrap beacon.
	IsBeacon bool
	// Must not be nil.
	StakingKey []byte
	// Must not be nil.
	StakingCert []byte
	// May be nil.
	ConfigFile []byte
	// May be nil.
	CChainConfigFile []byte
}

// Returns an error if this config is invalid
func (c *Config) Validate(expectedNetworkID uint32) error {
	switch {
	case c.ImplSpecificConfig == nil:
		return errors.New("implementation-specific node config not given")
	case len(c.StakingKey) == 0:
		return errors.New("staking key not given")
	case len(c.StakingCert) == 0:
		return errors.New("staking cert not given")
	default:
		return validateConfigFile(c.ConfigFile, expectedNetworkID)
	}
}

// Returns an error if config file [configFile] is invalid.
// If len([configFile]) == 0, returns nil.
func validateConfigFile(configFile []byte, expectedNetworkID uint32) error {
	if len(configFile) == 0 {
		// No config file given. Skip.
		return nil
	}
	// Validate that values given in the config file,
	// if any, are the correct type.
	var configMap map[string]interface{}
	if err := json.Unmarshal(configFile, &configMap); err != nil {
		return fmt.Errorf("could not unmarshal config file: %w", err)
	}
	if networkIDIntf, ok := configMap[config.NetworkNameKey]; ok {
		networkID, ok := networkIDIntf.(float64)
		if !ok {
			return fmt.Errorf("wrong type for field %q in config expected float64 got %T", config.NetworkNameKey, networkIDIntf)
		}
		if uint32(networkID) != expectedNetworkID {
			return fmt.Errorf("config file network id %d differs from genesis network id %d", int(networkID), expectedNetworkID)
		}
	}
	if dbPathIntf, ok := configMap[config.DBPathKey]; ok {
		if _, ok := dbPathIntf.(string); !ok {
			return fmt.Errorf("wrong type for field %q in config expected string got %T", config.DBPathKey, dbPathIntf)
		}
	}
	if logsDirIntf, ok := configMap[config.LogsDirKey]; ok {
		if _, ok := logsDirIntf.(string); !ok {
			return fmt.Errorf("wrong type for field %q in config expected string got %T", config.LogsDirKey, logsDirIntf)
		}
	}
	if apiPortIntf, ok := configMap[config.HTTPPortKey]; ok {
		if _, ok := apiPortIntf.(float64); !ok {
			return fmt.Errorf("wrong type for field %q in config expected float64 got %T", config.HTTPPortKey, apiPortIntf)
		}
	}
	if p2pPortIntf, ok := configMap[config.StakingPortKey]; ok {
		if _, ok := p2pPortIntf.(float64); !ok {
			return fmt.Errorf("wrong type for field %q in config expected float64 got %T", config.StakingPortKey, p2pPortIntf)
		}
	}
	return nil
}

// An AvalancheGo node
type Node interface {
	// Return this node's name, which is unique
	// across all the nodes in its network.
	GetName() string
	// Return this node's Avalanche node ID.
	GetNodeID() ids.ShortID
	// Return a client that can be used to make API calls.
	GetAPIClient() api.Client
	// TODO add methods
}
