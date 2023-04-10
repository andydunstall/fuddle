package start

var (
	gossipBindAddr string
	gossipBindPort int
	gossipAdvAddr  string
	gossipAdvPort  int

	gossipSeeds string

	rpcBindAddr string
	rpcBindPort int
	rpcAdvAddr  string
	rpcAdvPort  int

	logLevel string
)

func init() {
	Command.Flags().StringVarP(
		&gossipBindAddr,
		"gossip-bind-addr", "",
		"0.0.0.0",
		"the bind address to listen for gossip traffic",
	)
	Command.Flags().IntVarP(
		&gossipBindPort,
		"gossip-bind-port", "",
		8111,
		"the bind port to listen for gossip traffic",
	)

	Command.Flags().StringVarP(
		&gossipAdvAddr,
		"gossip-adv-addr", "",
		"",
		"the advertised address for gossip traffic (defaults to the bind addr)",
	)
	Command.Flags().IntVarP(
		&gossipAdvPort,
		"gossip-adv-port", "",
		0,
		"the advertised port for gossip traffic (defaults to the bind addr)",
	)

	Command.Flags().StringVarP(
		&gossipSeeds,
		"join", "",
		"",
		"gossip addresses in the target cluster to join",
	)

	Command.Flags().StringVarP(
		&rpcBindAddr,
		"rpc-bind-addr", "",
		"0.0.0.0",
		"the bind address to listen for rpc traffic",
	)
	Command.Flags().IntVarP(
		&rpcBindPort,
		"rpc-bind-port", "",
		8110,
		"the bind port to listen for rpc traffic",
	)

	Command.Flags().StringVarP(
		&rpcAdvAddr,
		"rpc-adv-addr", "",
		"",
		"the advertised address for rpc traffic (defaults to the bind addr)",
	)
	Command.Flags().IntVarP(
		&rpcAdvPort,
		"rpc-adv-port", "",
		0,
		"the advertised port for rpc traffic (defaults to the bind addr)",
	)

	Command.Flags().StringVarP(
		&logLevel,
		"log-level", "",
		"info",
		"the log level to use (one of 'debug', 'info', 'warn', 'error')",
	)
}
