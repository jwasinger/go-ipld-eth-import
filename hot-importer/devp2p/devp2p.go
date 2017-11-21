package devp2p

import (
  "bufio"
  "fmt"
  "os"

  "github.com/ethereum/go-ethereum/log"
  "github.com/ethereum/go-ethereum/p2p"
  "github.com/ethereum/go-ethereum/p2p/discover"
  //"github.com/ethereum/go-ethereum/ethdb"
  //"github.com/ethereum/go-ethereum/core"
  colorable "github.com/mattn/go-colorable"
)

// DevP2P is the object that organizes and enables the connection
// of a node to the devp2p network, it can start and stop a server
// and make available data and metrics through an API.
type DevP2P struct {
  server *p2p.Server
  pm *p2p.ProtocolManager
  chainDb ethDb.Database
  blockchain *core.BlockChain
  engine consensus.Engine
  protoManager *ProtocolManager
}

// Config is the configuration object for DevP2P
type Config struct {
  // bootnodes file
  BootnodesPath string

  // bootnodes slice
  bootnodes []*discover.Node

  // node database path. Must be appointed outside this package
  NodeDatabasePath string

  // we can find the client's private key here
  PrivateKeyFilePath string

  // glogger verbosity level (5 is the highest)
  Verbosity int

  // glogger verbosity per module (ex: devp2p=5,p2p=5)
  Vmodule string
}

func OpenDatabase(path string, cache int, handles int) (ethdb.Database, error) {
  db, err := ethdb.NewLDBDatabase(path, cache, handles)
  if err != nil {
    return nil, err
  }
  return db, nil
}

// NewDevP2P returns a DevP2P Manager object
//
// * defines logger.
//
// * defines bootnodes.
//
// * defines node database.
//
// * sets up the _peerstore_.
//
// * defines and configures _server_, passing the _protocol-handler_.
// * _protocol-handler_ needs the _peer-status-msg_ methods to perform
// * an eth handshake, adding and removing peers from the _peerstore_.
// * also, _protocol-handler_ will put some received requests into channels,
// * we may want to answer them.

//
// * sets up _api_, this will talk to the _peer-send_ methods,
// * which in turn picks the best peers and ask the question.
//
// * sets up the _metrics_.
//
func NewDevP2P(config Config) *DevP2P, error {
  var err error

  setupLogger(config)

  config.bootnodes, err = parseBootnodesFile(config.BootnodesPath)
  if err != nil {
    log.Error("NewManager", "error", fmt.Sprintf("processBootnodesFile error: %v", err))
    os.Exit(1) // zero tolerance
  }

  if config.NodeDatabasePath == "" {
    log.Error("NewManager", "error", "node database path must be appointed outside this package")
    os.Exit(1)
  }

  path := "/home/jwasinger/db/data"

  eventMux = new(event.TypeMux)
  engine := ethash.NewFaker()

  chainDb, err := OpenDatabase(path, 128, nil)
  if err != nil {
    return nil, err
  }

  chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, nil /* use default main-net genesis */)
  if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
    return nil, genesisErr
  }

  log.Info("Initialised chain configuration", "config", chainConfig)

  //assume fast sync, full client
  vmConfig := vm.Config{EnablePreimageRecording: false} //dunno what this is.. probably should figure out :)

  blockchain, err = core.NewBlockChain(chainDb, chainConfig, engine, vmConfig)
  if err != nil {
    return nil, err
  }

  txPool = core.NewTxPool(config.TxPool, eth.chainConfig, eth.blockchain)

  if protocolManager, err = NewProtocolManager(chainConfig, downloader.LightSync, 1/*NetworkId*/, eventMux, txPool, engine, blockchain, chainDb); err != nil {
    return nil, err
  }

  server := newServer(config, protocolManager.SubProtocols)

  manager := &devp2p{
    blockchain: blockchain,
    engine: engine,
    chainDb: chainDb,
    chainConfig: chainConfig,
    server: server
  }

  return manager
}

func (d *DevP2P) Start() {
  maxPeers := d.server.MaxPeers

  if err := d.server.Start(); err != nil {
    log.Error("error starting devp2p server", "error", err)
    os.Exit(1)
  }

  if err := d.protoManager.Start(maxPeers); err != nil {
    log.Error("error starting ETH protocol manager", "error", err)
    os.Exit(1)
  }
}

// Stop terminates the server
func (d *DevP2P) Stop() {
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.eventMux.Stop()
	s.chainDb.Close()
  d.server.Stop()
}

// setupLogger configures glogger with the required verbosity.
func setupLogger(config Config) {
  output := colorable.NewColorableStderr()
  glogger := log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(true)))

  glogger.Verbosity(log.Lvl(config.Verbosity))
  glogger.Vmodule(config.Vmodule)

  log.Root().SetHandler(glogger)
}

// parseBootnodesFile parses the bootnodes file to be included in the
// devp2p server.
func parseBootnodesFile(filePath string) ([]*discover.Node, error) {
  nodes := []*discover.Node{}

  if filePath == "" {
    return nil, fmt.Errorf("A bootnodes file must be defined!")
  }

  file, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    nodeUrl := scanner.Text()
    node, err := discover.ParseNode(nodeUrl)
    if err != nil {
      log.Error("add bootstrap node error", "node-url", nodeUrl, "error", err)
    }
    nodes = append(nodes, node)
    log.Debug("added bootstrap node", "node-url", nodeUrl)
  }

  if err := scanner.Err(); err != nil {
    return nil, err
  }

  return nodes, nil
}
