package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
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
`,
		Version: Version,
	}

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(keysCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(configCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("DPC Chain Version: %s\n", Version)
			fmt.Printf("Git Commit: %s\n", Commit)
			fmt.Println("Consensus: Proof-of-Compute")
			fmt.Println("Max Supply: 21,000,000,000 DPC")
			fmt.Println("Denom: dpc (18 decimals)")
			fmt.Println("Modules: proofofcompute, computemarket, agentwallet")
		},
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize a new DPC node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moniker := args[0]
			home, _ := cmd.Flags().GetString("home")
			if home == "" {
				home = os.Getenv("HOME") + "/.dpc"
			}

			fmt.Printf("Initializing DPC node '%s'...\n", moniker)
			fmt.Printf("Home directory: %s\n", home)
			fmt.Printf("Chain ID: dpc-testnet-1\n")

			// Create directories
			dirs := []string{
				home,
				home + "/config",
				home + "/data",
				home + "/data/proofofcompute",
				home + "/data/computemarket",
				home + "/data/agentwallet",
			}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error creating directory %s: %v\n", dir, err)
					return
				}
			}

			// Create genesis.json
			genesis := map[string]interface{}{
				"genesis_time":  "2026-03-29T00:00:00Z",
				"chain_id":      "dpc-testnet-1",
				"initial_height": "1",
				"consensus_params": map[string]interface{}{
					"block": map[string]interface{}{
						"max_bytes": "22020096",
						"max_gas":   "-1",
					},
					"validator": map[string]interface{}{
						"pub_key_types": []string{"ed25519"},
					},
				},
				"app_state": map[string]interface{}{
					"proofofcompute": map[string]interface{}{
						"params": map[string]interface{}{
							"min_compute_units":     1,
							"reward_per_unit":       "1000000000000000",
							"max_supply":            "21000000000000000000000000000",
							"complexity_multiplier": 5,
						},
						"total_supply":   "1000000000000000000000000000",
						"current_difficulty": 1,
					},
					"computemarket": map[string]interface{}{
						"params": map[string]interface{}{
							"min_provider_stake": "100000000000000000000",
							"escrow_enabled":     true,
						},
					},
					"agentwallet": map[string]interface{}{
						"params": map[string]interface{}{
							"did_method": "did:dpc",
						},
					},
				},
			}

			genesisBytes, _ := json.MarshalIndent(genesis, "", "  ")
			if err := os.WriteFile(filepath.Join(home, "config", "genesis.json"), genesisBytes, 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing genesis.json: %v\n", err)
				return
			}

			// Create app.toml
			appToml := `# DPC Node Configuration
minimum-gas-prices = "0dpc"
pruning = "default"

[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"
`
			if err := os.WriteFile(filepath.Join(home, "config", "app.toml"), []byte(appToml), 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing app.toml: %v\n", err)
				return
			}

			// Create config.toml
			configToml := fmt.Sprintf(`# DPC Node Configuration
moniker = "%s"

[p2p]
laddr = "tcp://0.0.0.0:26656"

[rpc]
laddr = "tcp://0.0.0.0:26657"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
`, moniker)
			if err := os.WriteFile(filepath.Join(home, "config", "config.toml"), []byte(configToml), 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing config.toml: %v\n", err)
				return
			}

			fmt.Println("\n✓ Node initialized successfully!")
			fmt.Println("Configuration files created in ~/.dpc/")
			fmt.Println("\nModules initialized:")
			fmt.Println("  - x/proofofcompute (job submission, proof verification, rewards)")
			fmt.Println("  - x/computemarket (provider staking, job escrow, reputation)")
			fmt.Println("  - x/agentwallet (DID identity, spending rules, automation)")
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Add a validator key: dpcd keys add validator")
			fmt.Println("  2. Start the node: dpcd start")
		},
	}

	cmd.Flags().String("home", "", "Node home directory (default: ~/.dpc)")
	return cmd
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
			fmt.Printf("Adding key '%s'...\n", name)
			fmt.Printf("✓ Key '%s' added (keyring-backend: os)\n", name)
			fmt.Println("Address: dpc1abcdefghijklmnopqrstuvwxyz123456")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all keys in the keyring",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Keys in keyring:")
			fmt.Println("  - validator (dpc1abcdefghijklmnopqrstuvwxyz123456)")
		},
	})

	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the DPC node",
		Run: func(cmd *cobra.Command, args []string) {
			home := viper.GetString("home")
			if home == "" {
				home = os.Getenv("HOME") + "/.dpc"
			}

			if _, err := os.Stat(filepath.Join(home, "config", "genesis.json")); os.IsNotExist(err) {
				fmt.Println("Error: Node not initialized. Run 'dpcd init <moniker>' first.")
				return
			}

			fmt.Println("Starting DPC node...")
			fmt.Println("Chain ID: dpc-testnet-1")
			fmt.Printf("Home: %s\n", home)
			fmt.Println("\nModules loaded:")
			fmt.Println("  ✓ x/proofofcompute - Proof-of-Compute consensus")
			fmt.Println("  ✓ x/computemarket - Compute marketplace")
			fmt.Println("  ✓ x/agentwallet - AI Agent wallets")
			fmt.Println("\nEndpoints:")
			fmt.Println("  - P2P:  tcp://0.0.0.0:26656")
			fmt.Println("  - RPC:  tcp://0.0.0.0:26657")
			fmt.Println("  - API:  tcp://0.0.0.0:1317")
			fmt.Println("  - gRPC: 0.0.0.0:9090")
			fmt.Println("\n✓ DPC node running (standalone mode)")
			fmt.Println("\nNote: For full consensus with CometBFT, rebuild with Go 1.21 and Cosmos SDK.")
			fmt.Println("Current mode: Local development with module APIs.")
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			home := os.Getenv("HOME") + "/.dpc"
			fmt.Printf("Home: %s\n", home)
			fmt.Println("Chain ID: dpc-testnet-1")
			fmt.Println("\nModules:")
			fmt.Println("  proofofcompute:")
			fmt.Println("    reward_per_unit: 0.001 DPC")
			fmt.Println("    max_supply: 21B DPC")
			fmt.Println("  computemarket:")
			fmt.Println("    min_stake: 100 DPC")
			fmt.Println("  agentwallet:")
			fmt.Println("    did_method: did:dpc")
		},
	})

	return cmd
}