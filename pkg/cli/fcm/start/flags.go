package start

var (
	addr string
	port int

	cluster bool

	logLevel string

	clusterLogDir string
)

func init() {
	Command.Flags().StringVarP(
		&addr,
		"bind-addr", "",
		"127.0.0.1",
		"the address to listen on",
	)
	Command.Flags().IntVarP(
		&port,
		"bind-port", "",
		8220,
		"the port to listen on",
	)

	Command.Flags().BoolVarP(
		&cluster,
		"cluster", "",
		false,
		"whether to create a default cluster on startup",
	)

	Command.Flags().StringVarP(
		&logLevel,
		"log-level", "",
		"info",
		"the log level to use (one of 'debug', 'info', 'warn', 'error')",
	)

	Command.Flags().StringVarP(
		&clusterLogDir,
		"cluster-log-dir", "",
		"",
		"the log directory for cluster output (defaults to a temporary dir)",
	)
}
