package start

var (
	addr string
	port int

	logLevel string
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
		9110,
		"the port to listen on",
	)

	Command.Flags().StringVarP(
		&logLevel,
		"log-level", "",
		"info",
		"the log level to use (one of 'debug', 'info', 'warn', 'error')",
	)
}
