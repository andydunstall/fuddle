# RFC: Node Reconnect Support

**Status**: In Progress

Clients disconnecting from the registry and reconnecting will often occur,
either due to brief networking issues, rebalancing load or Fuddle servers
shedding connections. Therefore node clients reconnecting must not cause
disruption.

This RFC proposes:
* A method for client nodes to reconnect to Fuddle and guarantee they will not
miss any updates to the registry that occurred while they were disconnected
* A method for client nodes to retry connecting to Fuddle using exponential
backoff and jitter to avoid overloading the Fuddle servers

# Versioned Updates
Each entry in a registered nodes state is given a version number. Whenever the
state is updated (such as a metadata entry), the entry is given a version equal
to the largest version of all the nodes state plus one. The node itself has a
version equal to the maximum version of its state entries.

These versions are sent to the client, so when it reconnects, it can send the ID
and version of its known nodes to request any state it missed.

The server can use the ID-version pairs and send:
* All nodes that have been registered that the client doesn’t know about
* All updated node state for nodes the client does know about but is out of date
* All nodes that have been unregistered that the client believes are still
registered

Clients can only be disconnected for 30s without being forcefully
unregistered, so this delta of missed updates should be small.

If a client reconnects after being unregistered due to being considered failed,
the server will respond with an error telling the client it is not registered,
so the client will re-register itself.

Note the order of updates of a particular node matters, such as a `metadata`
update must come after a `register` update, though the order of updates of
different nodes doesn’t matter.

Node versions can also be used for detecting differences between Fuddle server
replicas when adding gossip.

# Registry API
Adds new message types to the registry API:

### `sync`
When a client reconnects it sends a `sync` message containing its node ID. This
also adds a new `sync` field containing a mapping of node ID to version of the
clients known nodes.

### `error`
The error message adds a new `error` field to the update message that contains a
status enum and error message. This is used when a node tries to reconnect by
sending a `sync`, though the node is not registered (likely due to missing too
many heartbeats), the server will respond with a `NOT_REGISTERED` error causing
the client to re-register itself.

# Reconnect Strategy
To avoid clients generating excessive load, they must use exponential backoff
with jitter when trying to reconnect to a Fuddle server.
