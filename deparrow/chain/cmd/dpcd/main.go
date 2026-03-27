package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version = "v1.0.0"
	Commit  = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dpcd",
		Short: "DEparrow Coin (DPC) blockchain daemon",
		Long: `DPC is the native cryptocurrency of the DEparrow Global Virtual Machine platform.
AI agents use DPC to autonomously buy compute and earn by providing services.

Features:
- Proof-of-Compute consensus (completed jobs = mining)
- Max supply: 21 billion DPC
- AI Agent autonomous wallets
- Integration with Bacalhau compute network

Note: This is a minimal build for testing. Full Cosmos SDK integration
requires Go 1.21 or lower due to slices.SortFunc API changes in Go 1.22+.
`,
		Version: Version,
	}

	// Add basic commands
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(keysCmd())
	rootCmd.AddCommand(startCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("DPC Chain Version: %s\n", Version)
			cmd.Printf("Git Commit: %s\n", Commit)
			cmd.Println("Consensus: Proof-of-Compute")
			cmd.Println("Max Supply: 21,000,000,000 DPC")
			cmd.Println("Denom: dpc (18 decimals)")
		},
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize a new DPC node",
		Long: `Initialize a new DPC node with the specified moniker.

This creates the ~/.dpc directory with:
- genesis.json - Initial chain state
- node_key.json - P2P node identity
- priv_validator_key.json - Validator signing key
- config/ - Node configuration files
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moniker := args[0]
			home := os.Getenv("HOME") + "/.dpc"
			
			cmd.Printf("Initializing DPC node '%s'...\n", moniker)
			cmd.Printf("Home directory: %s\n", home)
			cmd.Printf("Chain ID: dpc-testnet-1\n")
			
			// Create directories
			dirs := []string{
				home,
				home + "/config",
				home + "/data",
			}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					cmd.Printf("Error creating directory %s: %v\n", dir, err)
					return
				}
			}
			
			// Create genesis.json
			genesis := `{"genesis_time": "2026-03-28T00:00:00Z", "chain_id": "dpc-testnet-1", "initial_height": "1", "consensus_params": {"block": {"max_bytes": "22020096", "max_gas": "-1"}, "evidence": {"max_age_num_blocks": "100000", "max_age_duration": "172800000000000"}, "validator": {"pub_key_types": ["ed25519"]}}, "app_state": {"auth": {"params": {"max_memo_characters": "256", "tx_sig_limit": "7", "tx_size_cost_per_byte": "10", "sig_verify_cost_ed25519": "590", "sig_verify_cost_secp256k1": "1000"}}, "bank": {"params": {"send_enabled": true, "receive_enabled": true}, "balances": [], "supply": []}, "staking": {"params": {"unbonding_time": "1814400s", "max_validators": 100, "max_entries": 7, "historical_entries": 10000, "bond_denom": "dpc"}}, "mint": {"minter": {"inflation": "0.130000000000000000", "annual_provisions": "0.000000000000000000"}, "params": {"mint_denom": "dpc", "inflation_rate_change": "0.130000000000000000", "inflation_max": "0.200000000000000000", "inflation_min": "0.070000000000000000", "goal_bonded": "0.670000000000000000", "blocks_per_year": "6311520"}}}}`
			if err := os.WriteFile(home+"/config/genesis.json", []byte(genesis), 0644); err != nil {
				cmd.Printf("Error writing genesis.json: %v\n", err)
				return
			}
			
			// Create app.toml
			appToml := `# DPC Node Configuration
minimum-gas-prices = "0dpc"
pruning = "default"
pruning-keep-recent = "0"
pruning-keep-every = "0"
pruning-interval = "0"

[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"

[rosetta]
enable = false
`
			if err := os.WriteFile(home+"/config/app.toml", []byte(appToml), 0644); err != nil {
				cmd.Printf("Error writing app.toml: %v\n", err)
				return
			}
			
			// Create config.toml
			configToml := `# DPC CometBFT Configuration
moniker = "` + moniker + `"

[consensus]
create_empty_blocks = true
create_empty_blocks_interval = "0s"

[p2p]
laddr = "tcp://0.0.0.0:26656"

[rpc]
laddr = "tcp://0.0.0.0:26657"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
`
			if err := os.WriteFile(home+"/config/config.toml", []byte(configToml), 0644); err != nil {
				cmd.Printf("Error writing config.toml: %v\n", err)
				return
			}
			
			cmd.Println("\n✓ Node initialized successfully!")
			cmd.Println("Configuration files created in ~/.dpc/")
			cmd.Println("\nNext steps:")
			cmd.Println("  1. Add a validator key: dpcd keys add validator")
			cmd.Println("  2. Add genesis account: dpcd add-genesis-account <address> 1000000000dpc")
			cmd.Println("  3. Start the node: dpcd start")
		},
	}
}

func keysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage keyring and keys",
	}
	
	cmd.AddCommand(&cobra.Command{
		Use:   "add [name]",
		Short: "Add a new key to the keyring",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			cmd.Printf("Adding key '%s'...\n", name)
			cmd.Println("Note: Full keyring functionality requires Cosmos SDK build.")
			cmd.Println("This is a minimal build. For full functionality, rebuild with Go 1.21.")
			cmd.Printf("\nSimulated key '%s' added (keyring-backend: test)\n", name)
		},
	})
	
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all keys in the keyring",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Keys in keyring (keyring-backend: test):")
			cmd.Println("(No keys found - this is a minimal build)")
		},
	})
	
	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the DPC node",
		Run: func(cmd *cobra.Command, args []string) {
			home := os.Getenv("HOME") + "/.dpc"
			
			// Check if node is initialized
			if _, err := os.Stat(home + "/config/genesis.json"); os.IsNotExist(err) {
				cmd.Println("Error: Node not initialized. Run 'dpcd init <moniker>' first.")
				return
			}
			
			cmd.Println("Starting DPC node...")
			cmd.Println("Chain ID: dpc-testnet-1")
			cmd.Printf("Home: %s\n", home)
			cmd.Println("\nNote: Full consensus engine requires Cosmos SDK build.")
			cmd.Println("This is a minimal build for testing purposes.")
			cmd.Println("\nNode configuration:")
			cmd.Println("  - P2P: tcp://0.0.0.0:26656")
			cmd.Println("  - RPC: tcp://0.0.0.0:26657")
			cmd.Println("  - API: tcp://0.0.0.0:1317")
			cmd.Println("  - gRPC: 0.0.0.0:9090")
			cmd.Println("\n✓ Configuration verified. Ready for full build with Go 1.21.")
		},
	}
}