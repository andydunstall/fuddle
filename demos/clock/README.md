# Clock Service Demo
The clock service is a simple demo of using Fuddle to load balance requests
among a set of nodes.

The cluster contains two types of nodes:
* Clock service: Exposes a gRPC API that returns the current time
* Frontends: Exposes a REST API `GET /time` that loads balances requests among
the set of clock nodes

Although this is a trivial example, it shows how Fuddle can be used to register
members, lookup members matching a filter and subscribe to updates when the
registry membership changes.

The frontends use a custom gRPC resolver that resolves the clock service
addresses from Fuddle.

To simulate a real production cluster, the demo cluster periodically replaces
a node (including both Fuddle and application nodes). The frontend subscribes to
registry updates, so when the nodes in the cluster changes, it updates the gRPC
resolver with the new addresses.

Run the demo using `fuddle demo clock`.
