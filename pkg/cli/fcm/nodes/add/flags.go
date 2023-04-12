package add

var (
	clusterID string

	fuddleNodes int
	clientNodes int

	addr string
)

func init() {
	Command.Flags().StringVarP(
		&clusterID,
		"cluster-id", "",
		"",
		"cluster id",
	)
	// nolint
	Command.MarkFlagRequired("cluster-id")

	Command.Flags().IntVarP(
		&fuddleNodes,
		"fuddle-nodes", "",
		1,
		"number of Fuddle nodes to add",
	)
	Command.Flags().IntVarP(
		&clientNodes,
		"client-nodes", "",
		1,
		"number of client nodes to add",
	)

	Command.Flags().StringVarP(
		&addr,
		"addr", "",
		"localhost:8220",
		"fcm server address",
	)
}
