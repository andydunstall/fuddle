package config

type Registry struct {
	// Address to bind to and listen on. Used for both UDP and TCP gossip.
	BindAddr string
	BindPort int

	// Address to advertise to other cluster members.
	AdvAddr string
	AdvPort int
}
