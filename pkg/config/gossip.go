package config

type Gossip struct {
	// Address to bind to and listen on. Used for both UDP and TCP gossip.
	BindAddr string
	BindPort int

	// Address to advertise to other cluster members.
	AdvAddr string
	AdvPort int

	// Seeds contains a list of gossip addresses of nodes in the target cluster
	// to join.
	Seeds []string
}
