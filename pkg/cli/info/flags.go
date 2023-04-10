package info

var (
	// addr is the Fuddle registry server to query.
	addr string
)

func init() {
	Command.PersistentFlags().StringVarP(
		&addr,
		"addr", "a",
		"localhost:8110",
		"address of the Fuddle server to query",
	)
}
