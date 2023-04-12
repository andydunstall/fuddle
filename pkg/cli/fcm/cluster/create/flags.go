package create

var (
	fuddleNodes int
	clientNodes int

	addr string
)

func init() {
	Command.Flags().IntVarP(
		&fuddleNodes,
		"fuddle-nodes", "",
		3,
		"number of Fuddle nodes in the cluster",
	)
	Command.Flags().IntVarP(
		&clientNodes,
		"client-nodes", "",
		10,
		"number of client nodes in the cluster",
	)

	Command.Flags().StringVarP(
		&addr,
		"addr", "",
		"localhost:8220",
		"fcm server address",
	)
}
