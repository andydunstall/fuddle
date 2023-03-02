# Examples

These examples provide services that use Fuddle to manage the cluster, which
can be run using `fuddle demo`.

## [Random Number Service](./random)
The random number service provides a toy example showing a simple use of Fuddle.

The cluster includes two types of node:
* Frontends: Accept client requests and forward to the appropriate backend node,
* Random: Generates a random number

Run the service using `fuddle demo random`.

<p align="center">
  <img src='../assets/images/random-demo.png?raw=true' width='60%'>
</p>

## [Messaging Service](./messaging-service)
The messaging service aims to provide a more real world example using Fuddle.

This is a publish/subscribe service where clients publish and subscribe to channels. Those channels are partitioned among the set of nodes in the cluster using consistent hashing.

All nodes are symmetrical so clients can connect to a node at random and it will act as the request coordinator and publish/subscribe to the channel primary node.

Fuddle is used to:
* Lookup the appropriate node for a channel,
* Receive updates about the cluster state and rebalance the channels as needed, such as when a node joins or leaves the cluster

<p align="center">
  <img src='../assets/images/messaging-service.png?raw=true' width='60%'>
</p>
