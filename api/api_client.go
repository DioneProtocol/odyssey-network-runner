package api

import (
	"fmt"

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

// interface compliance
var (
	_ Client        = (*APIClient)(nil)
	_ NewAPIClientF = NewAPIClient
)

// APIClient gives access to most odysseygo apis (or suitable wrappers)
type APIClient struct {
	omega        omegavm.Client
	aChain       alpha.Client
	aChainWallet alpha.WalletClient
	dChain       delta.Client
	dChainEth    EthClient
	info         info.Client
	health       health.Client
	ipcs         ipcs.Client
	keystore     keystore.Client
	admin        admin.Client
	oindex       indexer.Client
	dindex       indexer.Client
}

// Returns a new API client for a node at [ipAddr]:[port].
type NewAPIClientF func(ipAddr string, port uint16) Client

// NewAPIClient initialize most of odysseygo apis
func NewAPIClient(ipAddr string, port uint16) Client {
	uri := fmt.Sprintf("http://%s:%d", ipAddr, port)
	return &APIClient{
		omega:        omegavm.NewClient(uri),
		aChain:       alpha.NewClient(uri, "A"),
		aChainWallet: alpha.NewWalletClient(uri, "A"),
		dChain:       delta.NewDChainClient(uri),
		dChainEth:    NewEthClient(ipAddr, uint(port)), // wrapper over ethclient.Client
		info:         info.NewClient(uri),
		health:       health.NewClient(uri),
		ipcs:         ipcs.NewClient(uri),
		keystore:     keystore.NewClient(uri),
		admin:        admin.NewClient(uri),
		oindex:       indexer.NewClient(uri + "/ext/index/O/block"),
		dindex:       indexer.NewClient(uri + "/ext/index/D/block"),
	}
}

func (c APIClient) OChainAPI() omegavm.Client {
	return c.omega
}

func (c APIClient) AChainAPI() alpha.Client {
	return c.aChain
}

func (c APIClient) AChainWalletAPI() alpha.WalletClient {
	return c.aChainWallet
}

func (c APIClient) DChainAPI() delta.Client {
	return c.dChain
}

func (c APIClient) DChainEthAPI() EthClient {
	return c.dChainEth
}

func (c APIClient) InfoAPI() info.Client {
	return c.info
}

func (c APIClient) HealthAPI() health.Client {
	return c.health
}

func (c APIClient) IpcsAPI() ipcs.Client {
	return c.ipcs
}

func (c APIClient) KeystoreAPI() keystore.Client {
	return c.keystore
}

func (c APIClient) AdminAPI() admin.Client {
	return c.admin
}

func (c APIClient) OChainIndexAPI() indexer.Client {
	return c.oindex
}

func (c APIClient) DChainIndexAPI() indexer.Client {
	return c.dindex
}
