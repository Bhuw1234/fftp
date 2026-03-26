package main

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/deparrow/dpc/app"
	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	tmlog "github.com/cometbft/cometbft/libs/log"
	dbm "github.com/cosmos/cosmos-db"
)

// NewRootCmd creates a new root command for dpcd.
func NewRootCmd() *cobra.Command {
	encodingConfig := app.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("") // In simapp, we don't use any prefix for env variables.

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
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			tmCfg := tmcfg.DefaultConfig()
			tmCfg.MinGasPrices = "0dpc"
			tmCfg.API.Enable = true
			tmCfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
			tmCfg.P2P.ListenAddress = "tcp://0.0.0.0:26656"
			tmCfg.Instrumentation.Prometheus = true
			tmCfg.Instrumentation.PrometheusListenAddr = ":26660"

			customTMConfig := func() servertypes.TMConfig {
				return tmCfg
			}

			customAppConfig := func() servertypes.AppConfig {
				cfg := serverconfig.DefaultConfig()
				cfg.API.Enable = true
				cfg.API.Address = "tcp://0.0.0.0:1317"
				cfg.GRPC.Enable = true
				cfg.GRPC.Address = "0.0.0.0:9090"
				return cfg
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", customTMConfig, customAppConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd
}

// initRootCmd initializes the root command.
func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	gentxModule := app.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, gentxModule.GenTxValidator, encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		config.Cmd(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, func(startCmd *cobra.Command) {
		AddStartFlags(startCmd)
	})

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(app.DefaultNodeHome),
	)

	// add rosetta commands
	rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler))
}

// queryCommand returns the sub-command to send queries to the app.
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		authcmd.GetDecodeCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetTxsByHeightCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// txCommand returns the sub-command to send transactions to the app.
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetTxCmd(),
		authcmd.GetFeeGrantTxCmd(),
	)

	app.ModuleBasics.AddTxCommands(cmd)

	// Workaround for issue: https://github.com/spf13/cobra/issues/1262
	// We need to explicitly add the flags to the command
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// newApp creates a new DPC application.
func newApp(
	logger tmlog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.NewDPCApp(
		logger,
		db,
		traceStore,
		true,
		appOpts,
		baseappOptions...,
	)
}

// appExport creates a new DPC application for export.
func appExport(
	logger tmlog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var dpcApp *app.DPCApp

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home is not set")
	}

	loadLatest := height == -1
	dpcApp = app.NewDPCApp(
		logger,
		db,
		traceStore,
		loadLatest,
		appOpts,
	)

	if height != -1 {
		if err := dpcApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return dpcApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			decoder := clientCtx.AddressCodec()

			return genutilcli.AddGenesisAccount(clientCtx, cmd, args, defaultNodeHome, decoder)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, "", "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(tmcli.OutputFlag, "text", "Output format (text|json)")

	return cmd
}

// AddStartFlags adds start flags to the start command.
func AddStartFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("trace-store", false, "Enable KVStore tracing")
	cmd.Flags().String("inv-check-period", "", "Assert registered invariants every N blocks")
}

var (
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)
