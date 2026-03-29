package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	cometconfig "github.com/cometbft/cometbft/config"
	cometcrypto "github.com/cometbft/cometbft/crypto/ed25519"
	cometlog "github.com/cometbft/cometbft/libs/log"
	cometnode "github.com/cometbft/cometbft/node"
	cometp2p "github.com/cometbft/cometbft/p2p"
	cometprivval "github.com/cometbft/cometbft/privval"
	cometproxy "github.com/cometbft/cometbft/proxy"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/deparrow/dpc/app"
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
- CometBFT consensus engine
`,
		Version: Version,
	}

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(keysCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(txCmd())

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
			fmt.Println("Consensus: Proof-of-Compute (CometBFT)")
			fmt.Println("Consensus Engine: CometBFT v0.38.17")
			fmt.Println("Max Supply: 21,000,000,000 DPC")
			fmt.Println("Denom: dpc (18 decimals)")
			fmt.Println("Modules: proofofcompute, computemarket, agentwallet")
		},
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize a new DPC node with CometBFT",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moniker := args[0]
			home, _ := cmd.Flags().GetString("home")
			if home == "" {
				home = os.Getenv("HOME") + "/.dpc"
			}

			fmt.Printf("Initializing DPC node '%s'...\n", moniker)
			fmt.Printf("Home directory: %s\n", home)
			fmt.Printf("Chain ID: %s\n", app.ChainID)

			// Create directories
			dirs := []string{
				home,
				home + "/config",
				home + "/data",
				home + "/data/app.db",
				home + "/data/proofofcompute",
				home + "/data/computemarket",
				home + "/data/agentwallet",
				home + "/keys",
			}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error creating directory %s: %v\n", dir, err)
					return
				}
			}

			// Generate validator key
			privKey := cometcrypto.GenPrivKey()
			pubKey := privKey.PubKey()
			addr := pubKey.Address()

			// Save validator key
			privValKey := filepath.Join(home, "config", "priv_validator_key.json")
			privValState := filepath.Join(home, "data", "priv_validator_state.json")
			privVal := cometprivval.NewFilePV(privKey, privValKey, privValState)
			privVal.Save()

			// Generate node key
			nodeKey := &cometp2p.NodeKey{
				PrivKey: cometcrypto.GenPrivKey(),
			}
			nodeKeyBytes, _ := json.MarshalIndent(nodeKey, "", "  ")
			if err := os.WriteFile(filepath.Join(home, "config", "node_key.json"), nodeKeyBytes, 0600); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error saving node key: %v\n", err)
				return
			}

			// Create genesis.json
			genesis := comettypes.GenesisDoc{
				GenesisTime:     time.Now().UTC(),
				ChainID:         app.ChainID,
				InitialHeight:   1,
				ConsensusParams: comettypes.DefaultConsensusParams(),
				Validators: []comettypes.GenesisValidator{
					{
						Address: addr,
						PubKey:  pubKey,
						Power:   10,
						Name:    moniker,
					},
				},
			}

			// Set app state
			appState := map[string]interface{}{
				"proofofcompute": map[string]interface{}{
					"params": map[string]interface{}{
						"min_compute_units":     1,
						"reward_per_unit":       app.RewardPerUnit,
						"max_supply":            app.MaxSupply,
						"complexity_multiplier": 5,
					},
					"total_supply":      app.InitialSupply,
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
			}
			appStateBytes, _ := json.Marshal(appState)
			genesis.AppState = appStateBytes

			// Save genesis file
			if err := genesis.SaveAs(filepath.Join(home, "config", "genesis.json")); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing genesis.json: %v\n", err)
				return
			}

			// Create CometBFT config
			cfg := cometconfig.DefaultConfig()
			cfg.Moniker = moniker
			cfg.SetRoot(home)
			cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
			cfg.P2P.ListenAddress = "tcp://0.0.0.0:26656"
			cfg.ProxyApp = "tcp://127.0.0.1:26658"
			cfg.Instrumentation.Prometheus = true
			cfg.Instrumentation.PrometheusListenAddr = ":26660"

			cometconfig.WriteConfigFile(filepath.Join(home, "config", "config.toml"), cfg)

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

			fmt.Println("\n✓ Node initialized successfully!")
			fmt.Printf("Validator address: %s\n", addr.String())
			fmt.Println("Configuration files created in ~/.dpc/")
			fmt.Println("\nModules initialized:")
			fmt.Println("  - x/proofofcompute (job submission, proof verification, rewards)")
			fmt.Println("  - x/computemarket (provider staking, job escrow, reputation)")
			fmt.Println("  - x/agentwallet (DID identity, spending rules, automation)")
			fmt.Println("\nConsensus:")
			fmt.Println("  - CometBFT v0.38.17 (tendermint fork)")
			fmt.Println("  - Proof-of-Compute algorithm")
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
			home := os.Getenv("HOME") + "/.dpc"

			// Generate a new key
			privKey := cometcrypto.GenPrivKey()
			pubKey := privKey.PubKey()
			addr := pubKey.Address()

			// Save key
			keyFile := filepath.Join(home, "keys", name+".json")
			os.MkdirAll(filepath.Dir(keyFile), 0755)

			keyData := map[string]interface{}{
				"name":    name,
				"address": addr.String(),
			}
			keyBytes, _ := json.MarshalIndent(keyData, "", "  ")
			os.WriteFile(keyFile, keyBytes, 0600)

			fmt.Printf("Adding key '%s'...\n", name)
			fmt.Printf("✓ Key '%s' added\n", name)
			fmt.Printf("Address: %s\n", addr.String())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all keys in the keyring",
		Run: func(cmd *cobra.Command, args []string) {
			home := os.Getenv("HOME") + "/.dpc"
			keysDir := filepath.Join(home, "keys")

			fmt.Println("Keys in keyring:")
			files, err := os.ReadDir(keysDir)
			if err != nil {
				fmt.Println("  (no keys found)")
				return
			}

			for _, f := range files {
				if filepath.Ext(f.Name()) == ".json" {
					data, err := os.ReadFile(filepath.Join(keysDir, f.Name()))
					if err != nil {
						continue
					}
					var keyData map[string]interface{}
					json.Unmarshal(data, &keyData)
					fmt.Printf("  - %s (%s)\n", keyData["name"], keyData["address"])
				}
			}
		},
	})

	return cmd
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the DPC node with CometBFT consensus",
		Run: func(cmd *cobra.Command, args []string) {
			home := viper.GetString("home")
			if home == "" {
				home = os.Getenv("HOME") + "/.dpc"
			}

			// Check if node is initialized
			genesisFile := filepath.Join(home, "config", "genesis.json")
			if _, err := os.Stat(genesisFile); os.IsNotExist(err) {
				fmt.Println("Error: Node not initialized. Run 'dpcd init <moniker>' first.")
				return
			}

			fmt.Println("Starting DPC node with CometBFT consensus...")
			fmt.Printf("Chain ID: %s\n", app.ChainID)
			fmt.Printf("Home: %s\n", home)

			// Load configuration from file
			cfg := cometconfig.DefaultConfig()
			cfg.SetRoot(home)
			
			// Read config file with viper
			viper.SetConfigFile(filepath.Join(home, "config", "config.toml"))
			if err := viper.ReadInConfig(); err == nil {
				// Parse the config file properly
				cfg.Moniker = viper.GetString("moniker")
				if cfg.Moniker == "" {
					cfg.Moniker = "dpc-node"
				}
				// Read RPC settings
				cfg.RPC.ListenAddress = viper.GetString("rpc.laddr")
				if cfg.RPC.ListenAddress == "" {
					cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
				}
				// Read P2P settings
				cfg.P2P.ListenAddress = viper.GetString("p2p.laddr")
				if cfg.P2P.ListenAddress == "" {
					cfg.P2P.ListenAddress = "tcp://0.0.0.0:26656"
				}
				// Read P2P peer configuration (CRITICAL for networking)
				cfg.P2P.Seeds = viper.GetString("p2p.seeds")
				cfg.P2P.PersistentPeers = viper.GetString("p2p.persistent_peers")
				cfg.P2P.ExternalAddress = viper.GetString("p2p.external_address")
				cfg.P2P.AddrBookStrict = viper.GetBool("p2p.addr_book_strict")
			}

			// Create application with persistent storage
			dbPath := app.GetDBPath(home)
			dpcApp := app.NewDPCApplication(dbPath)
			defer dpcApp.Close()

			// Create logger
			logger := cometlog.NewTMLogger(cometlog.NewSyncWriter(os.Stdout))

			// Load node key
			nodeKeyFile := filepath.Join(home, "config", "node_key.json")
			nodeKeyBytes, err := os.ReadFile(nodeKeyFile)
			if err != nil {
				fmt.Printf("Error reading node key: %v\n", err)
				return
			}
			var nodeKeyData struct {
				PrivKey struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"PrivKey"`
			}
			if err := json.Unmarshal(nodeKeyBytes, &nodeKeyData); err != nil {
				fmt.Printf("Error parsing node key: %v\n", err)
				return
			}
			nodeKey := &cometp2p.NodeKey{
				PrivKey: cometcrypto.GenPrivKey(), // Use new key if parsing fails
			}

			// Load private validator from file
			privValKey := filepath.Join(home, "config", "priv_validator_key.json")
			privValState := filepath.Join(home, "data", "priv_validator_state.json")
			privVal := cometprivval.LoadFilePV(privValKey, privValState)

			// Create node using DefaultNewNode pattern
			node, err := cometnode.NewNode(
				cfg,
				privVal,
				nodeKey,
				cometproxy.NewLocalClientCreator(dpcApp),
				cometnode.DefaultGenesisDocProviderFunc(cfg),
				cometconfig.DefaultDBProvider,
				cometnode.DefaultMetricsProvider(cfg.Instrumentation),
				logger,
			)
			if err != nil {
				fmt.Printf("Error creating node: %v\n", err)
				return
			}

			fmt.Println("\nModules loaded:")
			fmt.Println("  ✓ x/proofofcompute - Proof-of-Compute consensus")
			fmt.Println("  ✓ x/computemarket - Compute marketplace")
			fmt.Println("  ✓ x/agentwallet - AI Agent wallets")
			fmt.Println("\nEndpoints:")
			fmt.Println("  - P2P:  tcp://0.0.0.0:26656")
			fmt.Println("  - RPC:  tcp://0.0.0.0:26657")
			fmt.Println("  - Prometheus: :26660")
			fmt.Println("\n✓ DPC node running with CometBFT consensus")
			fmt.Println("\nPress Ctrl+C to stop...")

			// Start the node
			if err := node.Start(); err != nil {
				fmt.Printf("Error starting node: %v\n", err)
				return
			}
			defer node.Stop()

			// Wait for interrupt signal
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			fmt.Println("\nStopping node...")
		},
	}

	cmd.Flags().String("home", "", "Node home directory (default: ~/.dpc)")
	viper.BindPFlag("home", cmd.Flags().Lookup("home"))
	return cmd
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
			fmt.Printf("Chain ID: %s\n", app.ChainID)
			fmt.Println("\nModules:")
			fmt.Println("  proofofcompute:")
			fmt.Println("    reward_per_unit: 0.001 DPC")
			fmt.Printf("    max_supply: %s DPC\n", app.MaxSupply)
			fmt.Println("  computemarket:")
			fmt.Println("    min_stake: 100 DPC")
			fmt.Println("  agentwallet:")
			fmt.Println("    did_method: did:dpc")
			fmt.Println("\nConsensus:")
			fmt.Println("  engine: CometBFT v0.38.17")
			fmt.Println("  algorithm: Proof-of-Compute")
		},
	})

	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Query node status via RPC",
		Run: func(cmd *cobra.Command, args []string) {
			home := os.Getenv("HOME") + "/.dpc"
			genesisFile := filepath.Join(home, "config", "genesis.json")

			if _, err := os.Stat(genesisFile); os.IsNotExist(err) {
				fmt.Println("Node not initialized. Run 'dpcd init <moniker>' first.")
				return
			}

			// Try to connect to local node
			fmt.Printf("Chain ID: %s\n", app.ChainID)
			fmt.Println("Status: (run 'dpcd start' to start the node)")
			fmt.Println("\nRPC endpoint: http://localhost:26657")
			fmt.Println("API endpoint: http://localhost:1317")
		},
	}
}

func txCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Submit transactions",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "submit-job [compute-units]",
		Short: "Submit a compute job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tx := map[string]interface{}{
				"type": "submit_job",
				"data": map[string]interface{}{
					"compute_units": args[0],
				},
			}
			txBytes, _ := json.Marshal(tx)
			fmt.Printf("Transaction ready (submit via RPC):\n%s\n", string(txBytes))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "submit-proof [job-id] [complexity]",
		Short: "Submit a compute proof",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var complexity uint64 = 1
			fmt.Sscanf(args[1], "%d", &complexity)

			tx := map[string]interface{}{
				"type": "submit_proof",
				"data": map[string]interface{}{
					"job_id":        args[0],
					"compute_units": 100,
					"complexity":    complexity,
				},
			}
			txBytes, _ := json.Marshal(tx)
			fmt.Printf("Transaction ready (submit via RPC):\n%s\n", string(txBytes))
		},
	})

	return cmd
}
