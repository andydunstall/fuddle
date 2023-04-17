# Replication
:warning: In progress

In its simplest form, the registry is just a small distributed in-memory
database.

Since clients can connect to any Fuddle node to stream the registry state, the
registry must be replicated to every Fuddle node in the cluster.

Compared to a typical database, the number of members in the registry and the
rate of updates are both very low, so replicating member states to all nodes is
fine.

## Overview
Each member has its own stream to a node in Fuddle which is used to send updates
and heartbeats. The Fuddle node the members stream is connected to becomes the
owner for that member, so is responsible for propagating member updates and
verifying the member is healthy. When a member stream reconnects to another
Fuddle node, that node takes ownership of the member.

Every member in the registry is assigned a version by the owner, which can be
used to quickly detect missed updates and resolve conflicts when multiple nodes
believe they are the owner.

When an update is received for a member, or an update is generated by the
failure detector, the owner for the member forwards the update via RPC to all
other nodes in the cluster. Therefore under healthy conditions all Fuddle nodes
should quickly get all member updates.

To handle faults such as networking issues between nodes that could cause missed
updates, each node runs a background replica repair process. Each node will
periodically select a random node in the cluster, send it its known set of
member versions. The receiver will compare the member versions with its own
known versions and send any missed updates.