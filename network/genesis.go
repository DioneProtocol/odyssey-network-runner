package network

import (
	_ "embed"
	"encoding/json"
	"fmt"

	coreth_params "github.com/DioneProtocol/coreth/params"
)

//go:embed default/genesis.json
var genesisBytes []byte

// LoadLocalGenesis loads the local network genesis from disk
// and returns it as a map[string]interface{}
func LoadLocalGenesis() (map[string]interface{}, error) {
	var (
		genesisMap map[string]interface{}
		err        error
	)
	if err = json.Unmarshal(genesisBytes, &genesisMap); err != nil {
		return nil, err
	}

	dChainGenesis := genesisMap["dChainGenesis"]
	// set the dchain genesis directly from coreth
	// the whole of `dChainGenesis` should be set as a string, not a json object...
	corethDChainGenesis := coreth_params.OdysseyLocalChainConfig
	// but the part in coreth is only the "config" part.
	// In order to set it easily, first we get the dChainGenesis item
	// convert it to a map
	dChainGenesisMap, ok := dChainGenesis.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"expected field 'dChainGenesis' of genesisMap to be a map[string]interface{}, but it failed with type %T", dChainGenesis)
	}
	// set the `config` key to the actual coreth object
	dChainGenesisMap["config"] = corethDChainGenesis
	// and then marshal everything into a string
	configBytes, err := json.Marshal(dChainGenesisMap)
	if err != nil {
		return nil, err
	}
	// this way the whole of `dChainGenesis` is a properly escaped string
	genesisMap["dChainGenesis"] = string(configBytes)
	return genesisMap, nil
}
