# Replication
:warning: Note the registry is being replaced with [registry v2](../registryv2/registry.md).

The registry is replicated across all Fuddle nodes. This means clients can
connect to any node in the cluster and stream the entire registry state and all
registry updates.

# Ownership
Each member must have an assigned Fuddle node that acts as the authority for the
member. This Fuddle node is called the members owner.

The owner responsible for receiving member updates, replicating updates to other
nodes, and detecting if the member has failed.

## Assigning Ownership
As described in [registry.md](./registry.md), each member has an associated
client that registered the member.

The client connects to a Fuddle node in its region at random then sends updates
and heartbeats to the node. The connected Fuddle node becomes the owner for all
members the client registered.

The node updates the members entry in the registry to include an owner, where
the owner contains the ID of the node and the timestamp when the node became an
owner.

When clients reconnect to another Fuddle node, that node then becomes the owner
and updates the members owner field. The owner timestamp is used to detect
conflicts where multiple nodes believe they are the owner. When a node receives
an update about a member from another node, it will check whether the owner has
changed, and if it has, the node compares the owner timestamps of its known
owner and the new owner.

## Failure Detector
This builds upon the member failure detector described in
[registry.md](./registry.md).

### Moving Nodes
The owner of a member handles heartbeats with the connected client. If the
client disconnects for the `heartbeat_timeout` its members will be marked
`down`.

Though if the client reconnects to another node within this timeout, instead
another node will take ownership of the clients members and broadcast this
update to the other nodes. So the original owner will detect it is no longer the
owner and stop waiting for a heartbeat, so won’t mark the member as `down`.

### Fuddle Node Failure
If a Fuddle node goes down, its connected clients will reconnect to another node
which will take ownership of the clients members.

However it's possible that a client goes down at the same time as the connected
Fuddle node, so its ownership would never be updated, but the owner would never
mark the member as `down` or remove it.

Therefore when nodes detect another node as down or left (using gossip), they
will track the time that node left.

If the the the node has been gone exceeds the `heartbeat_timeout`, the other
nodes will try to take ownership of the member and mark it as `down`. Note
this will have nodes competing for ownership but eventually one will win.

So now that winning node has taken ownership, it will treat it as any other
down node that it owns and remove it if the member doesn't come back for the
`reconnect_timeout`.

If the node comes back, it will take back ownership as the member heartbeats
will give it the most up to date version, which will be replicated to other
nodes.

If the member reconnects to another node after the `heartbeat_timeout`, that
node will keep trying to take back ownership with every heartbeat and eventually
win.

# Streaming Updates
Each Fuddle node learns about other nodes in the cluster using gossip. Once a
node learns about another node, it will connect and stream all registry updates
to the set of members that node owns.

The target node starts by streaming any updates the source node has missed,
including:
* Register updates for new members owned by the target node
* State updates for member updates owned by the target
* Owner updates for members that the target no longer owns
* This is the only time a node responds with members that it doesn’t own, which
is preferred to sending an unregister as the source node could receive the
unregister before it learns about the new owner
* Unregister updates for members that are no longer owned by the target node but
don’t have a new owner

The source node will reject any updates whose owner timestamp is less than the
owner timestamp of the local copy of the member.

## Versioning
To detect missed updates, members are assigned a version.

The version includes:
* The owner ID
* The timestamp of the last update
* A counter to version updates in the same millisecond. The counter is
incremented for every update in the same millisecond, then reset to 0 in the
next millisecond

The version is updated whenever a member changes.

It can be used to detect that an owner has changed, and member updates from the
same owner.

## Tombstones
To assign a version to unregistered members, the left members are kept in the
registry with a status of `left` for the configured `tombstone_timeout` (default
30 minutes).

This must be greater than the `reconnect_timeout`, so that any partitioned node
that isn't getting unregister updates will unregister the member itself due to
losing contact with the members owner.

# Limitations

### Every Node Streams From Every Other Node
Every node must connect to and stream from every other node. This is fine in one
region, where the number of Fuddle nodes should be small, though it becomes an
issue where you have many Fuddle nodes distributed around multiple regions.

Therefore support for multiple regions is limited and will be improved in the
future.

### Partition Tolerance
Since each node gets all state members for another node directly from that node,
if there is a partition between two nodes they will miss updates from one
another even if both can still communicate indirectly.

Such as if you have a partition between nodes A and B, but both nodes can
communicate with C, then even though C knows all the members from A and B, A and
B still don’t know the members from each other.

In the future may improve this using a variant of Scuttlebutt so nodes can send
updates for any members, not just the members they own. So in the above example
node C could send node A any updates it's missing from node B.

### Requires Clock Synchronization
Given owner timestamps are used to determine the owner, clocks must be
synchronized. This is a common requirement, such as CockroachDB requires clock
synchronization, though may look at removing timestamps in the future. Such as
instead the client could track its own versions when it reconnects, so the owner
with the highest client version wins.

### Cluster Partition
Fuddle favors availability over consistency, so if the cluster is partitioned,
such as the network between availability zones goes down, you’ll end up with two
sides of the cluster who both believe the other side's nodes are down.
