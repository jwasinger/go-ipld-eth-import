package main

import "github.com/hermanjunge/devp2p-concept/devp2p"

func main() {
	// Get the flags
	cfg := ParseFlags()

	// Setup the devp2p server
	devp2pConfig := devp2p.Config{
		BootnodesPath:    cfg.BootnodesPath,
		NodeDatabasePath: cfg.NodeDatabasePath,
		Verbosity:        cfg.Verbosity,
		Vmodule:          cfg.Vmodule,
	}

	devp2pServer := devp2p.NewDevP2P(devp2pConfig)

	defer devp2pServer.Stop()
	// Start the devp2p server
	devp2pServer.Start()

	// Request for information to the devp2p network
	// TODO

	// PLACEHOLDER
	select {} // ... or a loop until we get the stuff we need
	// PLACEHOLDER

	// You got it? Print everything and good bye
	// TODO
}
