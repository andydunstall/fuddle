package cluster

var (
	nodes   int
	members int

	addr string
)

func init() {
	Command.Flags().IntVarP(
		&nodes,
		"nodes", "",
		3,
		"number of Fuddle nodes in the cluster",
	)
	Command.Flags().IntVarP(
		&members,
		"members", "",
		10,
		"number of registered random members in the cluster",
	)

	Command.Flags().StringVarP(
		&addr,
		"addr", "",
		"localhost:9110",
		"fcm server address",
	)
}
