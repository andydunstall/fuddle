# Examples

These examples provide services that use Fuddle to manage the cluster, which
can be run using `fuddle demo`.

## [Is Even Service](./is-even)
The ‘is even’ service provides a toy example showing a simple use of Fuddle.

The frontends expose a REST API to check whether a number is even or odd. When a request comes into the frontend, it queries both the ‘is-even’ service and the ‘is-odd’ service and returns the aggregated result.

Run the service using `fuddle demo is-even`.

<p align="center">
  <img src='../assets/images/is-even-service.png?raw=true' width='60%'>
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
