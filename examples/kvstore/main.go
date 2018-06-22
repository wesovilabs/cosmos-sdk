package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")

	rootDir := viper.GetString(cli.HomeFlag)
	db, err := dbm.NewGoLevelDB("basecoind", filepath.Join(rootDir, "data"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Capabilities key to access the main KVStore.
	var capKeyMainStore = sdk.NewKVStoreKey("main")

	// Create BaseApp.
	var baseApp = bam.NewBaseApp("kvstore", nil, logger, db)

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoresIAVL(capKeyMainStore)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	// Set a handler Route.
	baseApp.Router().AddRoute(NewHandler(capKeyMainStore))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(capKeyMainStore); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:26658", "socket", baseApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
	return
}

// KVStore Handler
type Handler struct {
	storeKey sdk.StoreKey
}

// NewHandler returns new handler
func NewHandler(storeKey sdk.StoreKey) sdk.Handler {
	return Handler{storeKey}
}

// Implements sdk.Handler
func (h Handler) Handle(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	dTx, ok := msg.(kvstoreTx)
	if !ok {
		panic("Handler should only receive kvstoreTx")
	}

	// tx is already unmarshalled
	key := dTx.key
	value := dTx.value

	store := ctx.KVStore(h.storeKey)
	store.Set(key, value)

	return sdk.Result{
		Code: 0,
		Log:  fmt.Sprintf("set %s=%s", key, value),
	}
}

// Implements sdk.Handler
func (h Handler) Type() string {
	return "kvstore"
}
