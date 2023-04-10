# Fuddle
> :warning: **Fuddle is still in development**

# What is Fuddle?
Fuddle is a service registry that can be used for client side service discovery
and cluster observability.

Unlike server side service discovery, where requests are simply load balanced
across a set of nodes in a service, client side service discovery gives
developers control. Using a Fuddle SDK nodes can register themselves, query the
set of registered members in the cluster and subscribe to changes.

Each registered member includes a set of attributes, including ID, service,
locality, etc, and application defined metadata.

## Features
* Service registry: Members register with the service registry to make
themselves known to the rest of the cluster
* Service discovery: Clients can lookup other members in the cluster, filter
based on service, locality and metadata, and subscribe to updates when the set
of matching members changes
* Cluster observability: The cluster and individual members can be inspected
using the Fuddle CLI

## Use Cases
Fuddle may be preferred to server side service discovery when you require
application specific routing, instead of load balancing among a set of stateless
service nodes.

Such as if you are using consistent hashing to partition work across multiple
nodes, you can use Fuddle to lookup nodes based on service, locality or
application defined metadata, and subscribe to updates to the matching set of
nodes to trigger a rebalance.

Or you may want a custom load balancing policy and retry strategy based on the
state of the nodes in the cluster. Such as routing based on node metadata, like
routing to the node with the lowest CPU.

# Getting Started
Start by installing the Fuddle server using
`go install github.com/fuddle-io/fuddle`.

## Server
Start the server using `fuddle start`.

# Docs
* Architecture
	* [Overview](./docs/architecture/overview.md)
	* [Registry](./docs/architecture/registry/registry.md)
		* [Replication](./docs/architecture/registry/replication.md)
		* [Client](./docs/architecture/registry/client.md)
