<p align="center">
  <img src='assets/images/logo.png?raw=true' width='60%'>
</p>

---

> :warning: **Fuddle is still in development**

---

# What is Fuddle?
Fuddle is a service registry that can be used for client side service discovery
and cluster observability.

Unlike server side service discovery, where requests are simply load balanced
across a set of nodes in a service, client side service discovery gives
developers control. Such as routing using consistent hashing or custom load
balancing policies.

Nodes register with Fuddle and include a set of fixed attributes (node ID,
service, locality etc) and application defined state. The set of nodes in the
cluster, and their state, is propagated to all nodes in the cluster. Nodes can
then lookup other nodes in the cluster and subscribe to updates.

## Features
* Service registry: Nodes register with the service registry to make themselves
known to the rest of the cluster
* Service discovery: Nodes can lookup other nodes in the cluster based on
service, locality or application defined state, and subscribe to updates when
the set of matching nodes changes
* Cluster observability: You can inspect the set of nodes in the cluster using
the Fuddle CLI

# Getting Started
Start by downloading the appropriate Fuddle binary for your platform from the
[releases](https://github.com/fuddle-io/fuddle/releases) page, or install
Fuddle using `go install github.com/fuddle-io/fuddle`.

The quickest way to start using Fuddle is to run a [demo](./demos) cluster:

### Counter Service
The counter service tracks the number of users subscribed with the same ID. So
users can register with ID `foo` and receive updates to the number of other
users subscribed with that ID as users join and leave.

Each counter service node in the cluster is responsible for a range of IDs using
consistent hashing. Fuddle is used to query the set of counter service nodes and
their addresses, and subscribe to updates when the set of nodes changes so needs
to rebalance.

The cluster can be started locally with `fuddle demo counter`.

See [demos/counter](./demos/counter) for details.

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
