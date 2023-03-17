<p align="center">
  <img src='assets/images/logo.png?raw=true' width='60%'>
</p>

---

> :warning: **Fuddle is still in development**

---

![CI Workflow](https://github.com/fuddle-io/fuddle/actions/workflows/main.yaml/badge.svg)

# What is Fuddle?
Fuddle is a service registry that can be used for client side service discovery
and cluster observability.

Unlike server side service discovery, where requests are simply load balanced
across a set of nodes in a service, client side service discovery gives
developers control. Using a Fuddle SDK nodes can register themselves, query the
set of nodes in the cluster and subscribe to changes.

Each registered node includes a set of fixed attributes (ID, service, locality
and revision), and application defined metadata which may be updated and
propagated to other nodes in the cluster.

## Features
* Service registry: Nodes register with the service registry to make themselves
known to the rest of the cluster
* Service discovery: Nodes can lookup other nodes in the cluster, filter based
on service, locality and metadata, and subscribe to updates when the set of
matching nodes changes
* Cluster observability: The cluster and individual nodes can be inspected using
the Fuddle CLI

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

## Demos
Once installed the quickest way to start playing around with Fuddle is to run
one of the [demo clusters](./demos) with `fuddle demo`:
* [Counter service](./demos/counter): An example of using Fuddle to implement
consistent hashing, where each node is responsible for a set of IDs, and counts
the number of users with the same ID. Run with `fuddle demo counter`

After a demo is running, it can be inspected using `fuddle status cluster` for a
cluster overview, or `fuddle status node {id}` to inspect a specific node in
detail.

## Server
To start using Fuddle in your own application, you can start the server with
`fuddle start`.

Note since this is an early version Fuddle only supports a single node so is not
recommended to be used in production (see details below).

## SDK
Use a Fuddle SDK to integrate Fuddle into your application, which lets nodes register themselves, discover other nodes in the cluster and subscribe to updates.

Only a [Go SDK](https://github.com/fuddle-io/fuddle-go) is supported so far.

# :warning: Limitations
Fuddle is still in early stages of development so has a number of limitations:

### No Fault Tolerance
Although the Fuddle API is defined, Fuddle can currently only run on a single
node. This means there is no fault tolerance of horizontal scaling, so is not
suitable for production use.

Also the Go SDK does not support reconnecting to the Fuddle node if the connection drops without potentially missing updates.

### Health Checks
Fuddle requires nodes to explicitly unregister when they are shutdown, so will
not detect nodes that have failed. This will be implemented using health checks.
