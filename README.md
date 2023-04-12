# Fuddle
> :warning: **Fuddle is still in development**

# What Is Fuddle?
Fuddle is a service registry, used for client side service discovery and cluster
observability.

Application nodes register themselves into the registry, then use the registry
to discover other nodes in the cluster, and information needed to interact with
those nodes.

Each member in the registry contains a set of attributes, including service,
locality and revision, plus metadata containing application-defined key-value
pairs. Clients can use the attributes and metadata to lookup members and
subscribe to changes.

# Design Goals

## Simplicity
Fuddle is built to be very simple to integrate into existing infrastructure.

Unlike many other service discovery systems, which require proxying requests and
running sidecars for every instance as part of a service mesh, Fuddle is a
standalone service that clients query via an SDK. 

## Flexibility
Fuddle supports client side service discovery instead of server side.

Server side discovery requires proxying requests via some load balancer, which
routes requests among a set of nodes registered for that service. This limits
how much control developers have over how they route requests and what
transports they can use.

Instead when using Fuddle, developers can query the registry using a Fuddle SDK
to lookup the target node(s).

This gives you more flexibility in how you route requests, such as:
* Consistent hashing: Using Fuddle to lookup a set of nodes that build a hash
ring, and subscribe to changes in the set of nodes to trigger a rebalance
* Custom load balancing: Instead of a simple round robin strategy, you can add
custom policies like weighted load balancing based on the target nodes metadata
* Transports: Since Fuddle is used just to look up the target nodes instead of
proxying requests, there's no constraints on what transports or protocols can be
used

Each node can also register custom metadata, so there's no limit to what
information can be shared with other nodes.

## Availability and Fault Tolerance
The registry is replicated over multiple Fuddle nodes, so if a node goes down,
clients automatically reconnect to another node without any disruption.

The registry is eventually consistent and favours availability over consistency.
It will also detect when members registered by users' applications go down and
removes them from the registry.

# Usage
Fuddle can be installed using `go install github.com/fuddle-io/fuddle`. This
includes a CLI to start a server node and interact with the cluster.

A Fuddle SDK can be used to register members and subscribe to the registry. So
far only a Go SDK ([fuddle-go](https://github.com/fuddle-io/fuddle-go)) is supported.

## Demo
The quickest way to get started with Fuddle is to run a demo cluster locally
using `fuddle demo`.

Such as `fuddle demo clock` will run a clock service cluster as described
[here](demos/clock/README.md).

## Start A Node
Start a Fuddle node with `fuddle start`. The node can be configured to join a
cluster using `--join`.

See `fuddle start –help` for details.

## Inspect A Cluster
`fuddle info` can be used to inspect all nodes in a cluster (including both
Fuddle nodes and members registered by the application).

`fuddle info cluster` lists all members of the cluster and their attributes.

`fuddle info member <id>` describes the member with the given ID, including
attributes and metadata.

# Documentation

## Usage
* [FCM](./docs/usage/fcm.md)

## Architecture
* [Overview](./docs/architecture/overview.md)
* [Registry](./docs/architecture/registry/registry.md)
	* [Replication](./docs/architecture/registry/replication.md)
	* [Client](./docs/architecture/registry/client.md)

# :warning: Limitations
Fuddle is still early in development so is missing features needed to run in
production.
