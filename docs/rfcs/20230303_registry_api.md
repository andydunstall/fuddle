# RFC: Registry API

**Status**: Done

The Fuddle registry maintains the set of nodes in the cluster and their state.

This RFC proposes an API for nodes to register themselves, update their state
and discover other nodes in the cluster.

# Requirements

## Must Include
* Nodes must register themselves when they join the cluster
* Nodes must unregister themselves when they leave the cluster
* Nodes must update the registry when their local state changes
* Nodes must receive the state of the other nodes in the cluster and
any updates to those nodes state

## Won't Include
Since this RFC proposes a first version of the registry API, it won't include:
* Handling node failures: Detecting when a node has failed and updating the
cluster
* Handling node disconnects: Supporting brief disconnects from the registry
by reconnecting without missing any state
* Filtering: Instead of receiving all state from all nodes in the cluster nodes
should query only the state their interested in

# Node State

## Attributes
Attributes contain immutable information about the node:
* Node ID (`string`): A unique identifier for the node in the cluster
* Service (`string`): The type of service running on the node
* Locality (`string`): The location of the node. Such as the availability zone
in a cloud deployment or a rack when hosted. This is recommended to be organised
into a hierarchy, such as `<provider>.<region>.<zone>` or
`<data center>.<rack>`, so make it easy to filter using globs. Such as
`aws.eu-*` to match all regions in AWS EU
* Created (`int64`): The UNIX timestamp in milliseconds that the node was
created
* Revision (`string`): An identifier for the version of the service running on
the node, such as a Git tag or commit SHA

## Metadata
Nodes include application defined metadata of type `map<string, string>`.

Unlike node attributes, the metadata may be updated.

Similar to the node locality, the keys are recommended to be organised into
a hierarchy to make filtering easier. Such as `addr.redis.hostname`, so a node
could query `addr.*`, or `addr.redis.*`.

(Note filtering isn't in the RFC but will be added later.)

# Cluster Lookup
Once a node registers, they receive the state of the cluster from the registry
and stream updates. Therefore each node maintains a local eventually consistent
view of the cluster.

This avoids having to do an RPC to Fuddle every time the node needs to lookup
state for another node. Instead they just query their local in-memory store.

# API
The registry exposes a REST API. This means the admin CLI and dashboard can
use the same API as the client SDKs. It also makes writing clients easier.

Messages are encoded with JSON, which should be ok for now, though can replace
with a more efficient encoding like msgpack in the future if needed.

When clients need a bidirectional stream, WebSockets are used.
WebSockets are used when clients need a bidirectional stream.

## Register: `/api/v1/register`
Register is used by clients to register themselves in the registry, and
stream updates to and from the registry.

### Message
Each update message (to and from the registry contains):
* Node ID (`string`): The registered nodes ID
* Update type (`string`): The type of update:
	* `register`: A node has joined the cluster
	* `unregister`: A node has left the cluster
	* `metadata`: A nodes metadata has changed
* Attributes: If the update is type `register` the nodes attributes are send. Since
the attributes are immutable, they are only sent in `register` updates
* Metadata: The nodes application defined metadata. The contents of this field
depend on the update type:
	* `register`: Contains all metadata fields
	* `unregister`: Empty since theres no need to update the metadata of a leaving
node
	* `metadata`: Contains only the key-value pairs that have been updated

### Stream
When the node first registers, it must send a `register` message containing its
own attributes and metadata.

The server will then send the node all existing nodes in the cluster as
`register` updates.

When a node updates its local metadata, it sends a `metadata` update to the
registry, which will broadcast the update to all other nodes.

Similarly when a node is shutdown it sends a `unregister` update to unregister
itself, which is again broadcast to all other nodes.

## Cluster: `/api/v1/cluster`
Returns the set of nodes in the cluster, including the attributes of each
node but no the node metadata.

## Nodes: `/api/v1/node/{id}`
Returns the attributes and metadata for the node with the given ID.

# Operating

## Metrics

### `fuddle_registry_node_count`
A gauge of the number of nodes in the cluster.

### `fuddle_registry_connection_count`
A gauge of the number of registry connections to this node.

### `fuddle_registry_update_count`
A counter of the number of registry updates received. With labels for the
update type (`register`, `unregister`, `metadata`).
