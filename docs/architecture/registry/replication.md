# Replication
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
mark the member as `down`.

Therefore when nodes detect another node as down (using gossip), they will wait
for the `heartbeat_timeout` for the member to get a new owner. If a member
doesn’t get a new owner in this time, it is marked as `down`, and eventually
unregistered if it doesn’t come back within the `reconnect_timeout`.

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
To detect missed updates, members are assigned a version number which starts at
1 and is incremented whenever the node's state changes.

The version assigned by the owner and reset whenever the owner changes, since
the new owner may not know the latest version from the previous owner.

When a node connects or reconnects to another node to stream updates, it
includes the set of versions and owners for its known nodes, so the target node
can determine what updates it's missing.

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
