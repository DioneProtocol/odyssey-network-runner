package api

import (
	"github.com/DioneProtocol/coreth/plugin/delta"
	"github.com/DioneProtocol/odysseygo/api/admin"
	"github.com/DioneProtocol/odysseygo/api/health"
	"github.com/DioneProtocol/odysseygo/api/info"
	"github.com/DioneProtocol/odysseygo/api/ipcs"
	"github.com/DioneProtocol/odysseygo/api/keystore"
	"github.com/DioneProtocol/odysseygo/indexer"
	"github.com/DioneProtocol/odysseygo/vms/alpha"
	"github.com/DioneProtocol/odysseygo/vms/omegavm"
)

// Issues API calls to a node
// TODO: byzantine api. check if appropriate. improve implementation.
type Client interface {
	OChainAPI() omegavm.Client
	AChainAPI() alpha.Client
	AChainWalletAPI() alpha.WalletClient
	DChainAPI() delta.Client
	DChainEthAPI() EthClient // ethclient websocket wrapper that adds mutexed calls, and lazy conn init (on first call)
	InfoAPI() info.Client
	HealthAPI() health.Client
	IpcsAPI() ipcs.Client
	KeystoreAPI() keystore.Client
	AdminAPI() admin.Client
	OChainIndexAPI() indexer.Client
	DChainIndexAPI() indexer.Client
	// TODO add methods
}
