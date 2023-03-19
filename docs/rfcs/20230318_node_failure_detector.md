# RFC: Node Failure Detector

**Status**: In Progress

This RFC proposes a way for Fuddle to detect failed nodes and remove them from
the registry.

To avoid confusion, using the term 'node' to refer to a node registered by the
client with Fuddle, and 'server' to refer to a Fuddle instance.

# Heartbeats
The Fuddle server detects when a registered node has failed using heartbeats.

Registered nodes must send a `PING` request every 10 seconds to the connected
Fuddle server. If the server doesn’t receive a `PING` for 30 seconds, it
unregisters the node.

The server responds with a `PONG` so the client can detect whether the server it
is connected to is down, and attempt to connect to another server if it is.

The `PING` request includes the client's current timestamp which is echoed back
in the `PONG` response, which the client can use to calculate the RTT with the
server.

Since clients can disconnect due to brief network disruption or connection
rebalancing, the clients connection is independent of failure detection.

# Operating

## Metrics

### `fuddle_registry_failure_count`
A counter of the number of nodes detected as failed.

# Limitations
This approach is very simple since this is an early version of Fuddle, though it
has some limitations that should be addressed in future versions:

## Cluster Disruption
Disruption in the cluster, such as networking issues or overlaoded Fuddle
servers unable to accept connections, could result in Fuddle incorrectly
unregistering a large number of client nodes.

Therefore may decide to rate limit how fast nodes can be unregistered due to
failing health checks, similar to [Eurekas](https://github.com/Netflix/eureka)
self preservation mode.

## Lack Of Control
This simple heartbeat strategy may be unsuitable for some use cases that require
more control over when a node is unregistered.

Such as when building a database, a node leaving the cluster could require a
large transfer of data to rebalance the cluster, so may prefer to limit how fast
nodes can be unregistered, or leave to an administrator to manually unregister
failed nodes.

Given one of the main aims of Fuddle is to give developers more control over the
cluster this should be supported in the future. Such as to notify the cluster
that a node is unreachable, then add configuration for how the down node should
be handle by Fuddle, such as:
* Unregister the node as soon as it is down,
* Mark the node as down but don't unregister for a configured window,
* Mark the node as down but leave it to an administrator to manually unregister
the node

These configuration options should also be configurable per application service
instead of global. Plus configuration for the heartbeat interval and number of
heartbeats missed before considering the node down.

## Lack Of Visibility
With failed nodes immediately being removed from the registry, this doesn’t give
administrators any chance to inspect the cluster to see which nodes are down.

Instead Fuddle should remember failed nodes for some interval and display them
in the admin cluster status.
