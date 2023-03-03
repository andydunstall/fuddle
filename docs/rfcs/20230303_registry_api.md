# RFC: Registry API

**Status**: In progress

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
* Revision (`string`): An identifier for the version of the service running on
the node, such as a Git tag or commit SHA

## State
Nodes include application defined state of type `map<string, string>`. This
state is propagated to the other nodes in the cluster.

Unlike node attributes, the state may be updated.

Similar to the node locality, the keys are recommended to be organised into
a hierarchy to make filtering easier. Such as `addr.redis.hostname`, so a node
could query `addr.*`, or `addr.redis.*`.

(Note filtering isn't proposed in the RFC but will be added later.)

# Cluster Lookup
Once a node registers, they receive the state of the cluster from the registry
and stream updates. Therefore each node maintains a local eventually consistent
view of the cluster.

This avoids having to do an RPC to Fuddle every time the node needs to lookup
state for another node. Instead they just query their local in-memory store.

# API

## Transport
Nodes communicate to Fuddle via gRPC. This makes it easy to define a schema and
generate code for SDKs in multiple languages. It also supports bidirectional
streaming which can be used to receive and send cluster state.

## Protocol
Each node opens a stream to Fuddle when it starts up. The stream is used by
the node to send its own state updates to the registry, and receive updates
about the rest of the cluster.

### Update Message
Each update (both to and from the registry) contains:
* Node ID (`string`): A unique identifier for the node being updated,
* Update type (`enum`): The type of update, which is either:
	* `JOIN`: A node has joined the cluster
	* `LEAVE`: A node has left the cluster
	* `UPDATE`: A nodes state has been updated
* Attributes: If the update is a `JOIN`, all the nodes attributes are sent.
The attributes are not included in `LEAVE` or `UPDATE` events since they are
immutable
* State: The nodes application state. The contents of this field depend on the
update type:
	* `JOIN`: All the nodes state must be included
	* `LEAVE`: Empty since theres no need to update the state of a node thats
left the cluster
	* `UPDATE`: Contains only the key-value pairs that have been updated

### Register
When a node registers it opens the stream and sends an update that includes its
node state with type `JOIN`.

Fuddle will then send it the states of the other nodes in the cluster as
`JOIN` updates, and continue streaming cluster update events when nodes join,
leave or update.

### Local Update
When the nodes local state is updated, it sends an update to the registry with
type `UPDATE` and includes the updated state entries.

### Unregister
When a node is shutdown it should send a `LEAVE` update to tell the registry
it has left the cluster.

Note as mentioned in the requirements, Fuddle doesn't yet support detecting node
failures.
