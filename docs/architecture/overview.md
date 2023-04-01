# Architecture Overview
This document contains a high level overview of the Fuddle architecture.

# Terminology
* Member: An entry in the registry identifying an instance of a service
* Registry: Contains the set of members registered
* Cluster: A group of Fuddle server nodes that maintain the registry
* Node: A Fuddle server node the participates in the cluster
* Client: A client running in a user process that streams registry updates and
registers members into the registry, typically using a Fuddle SDK

# Registry
The registry contains the set of registered members.

Clients connect to Fuddle, which they use to:
* Subscribe to registry updates, which can then be used for service discovery
and observability
* Register local members
* Update the state of their local members and unregister members

Each member registered member contains state:
* Attributes for lookup and observability, including the members ID, service,
locality, created time and revision
* Metadata containing arbitrary key-value pairs defined by the application
* A status of either `up` or `down` indicating whether Fuddle thinks the client
that registered the member is reachable

Each member is attached to the client that registered it. So if this client
stops responding, all members registered by the client are considered down.

Clients will usually be run as part of the application process being registered,
though could also run in a sidecar process, such as to register a 3rd party
service like Redis.

## Failure Detection
Each client with registered members must send regular heartbeats to Fuddle to be
considered active. If a client does not send a heartbeat for the configured
`heartbeat_timeout` (default to 30s), all members registered by that client are
marked as `down`.

If the client doesn’t come back within the configured `reconnect_timeout`
(defaults to 10 minutes), the members are unregistered and removed from the
registry, so if the client comes back after this timeout its members must be
re-registered.

# Clustering
The registry is distributed across a cluster of Fuddle nodes for fault tolerance
and scaling.

<p align="center">
  <img src='../../assets/images/cluster_overview.png?raw=true' width='60%'>
</p>

## Cluster Membership
Fuddle nodes must discover one another and detect node failures. This is done
with gossip using the [memberlist](https://pkg.go.dev/github.com/hashicorp/memberlist)
library. This provides eventually consistent cluster membership where updates
converge quickly.

## Replication
Each Fuddle node maintains a replica of the registry. The nodes will connect to
one another and stream the updates to the set of members owned by that node.

Streaming directly was preferred over gossiping each node's state. Gossip is
great for a large number of nodes, each with a small amount of state, though not
for streaming large amounts of data between a small number of nodes.

Not expecting to have many Fuddle nodes in the cluster, though the registry
may contain tens of thousands of members.
