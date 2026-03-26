package app

import (
	"encoding/json"
)

// GenesisState represents the genesis state of the blockchain.
// It is a map from module identifier strings to raw json messages.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	encCfg := MakeEncodingConfig()
	return ModuleBasics.DefaultGenesis(encCfg.Codec)
}
