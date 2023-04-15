package health

var (
	clusterID string

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

	Command.Flags().StringVarP(
		&addr,
		"addr", "",
		"localhost:8220",
		"fcm server address",
	)
}
