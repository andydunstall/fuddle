# Client
The Fuddle client is used by applications to subscribe to registry updates and
register local members. This document describes how clients interact with
Fuddle.

# Connection
When the client first connects to Fuddle, it is configured with a set of seed
nodes of Fuddle servers to try and connect to. It will try the nodes in random
order to balance load when multiple clients all have the same configured seed
addresses.

If the client can’t connect to any of the seeds, the connection will fail.

Clients are configured with:
* Seed addresses: This can either be a static list of addresses, or a callback
that fires whenever a new set of seeds is needed, which can be used if the list
is dynamic
* Connect timeout: The timeout to connect to any node. In Go this is implemented
as a context that can be cancelled by the user
* Connect attempt timeout (default 4s): The timeout for each connection attempt
to a particular node. So the total connect timeout should be longer than the
attempt timeout to give it time to try multiple nodes

## Reconnect
Once the client connects for the first time, it will automatically reconnect if
the connection drops.

Since the client gets a list of Fuddle nodes from its local registry, it can use
these node addresses. If none of the nodes from the registry works, it will fall
back to the seed addresses.

Clients use exponential backoff with jitter when retrying to avoid a large
number of clients all trying to reconnect at once and overloading the servers.

Clients are configured with:
* Initial backoff (default 500ms)
* Backoff multiplier (default 2)
* Maximum backoff (default 20s)
* Jitter multiplier (default 0.2)

The client will not stop retrying until it is closed by the user. If it tries
all Fuddle nodes without success, it will try them again.

There is an optional callback for the user to get notified when the connection
state changes (either connected or disconnected).

# Registry Subscription
Clients subscribe to a Fuddle nodes replica of the registry to build their own
eventually consistent view of the registry.

When the client first starts, it calls the `Subscribe` RPC stream and listens
for updates in the background. All updates are applied to its local registry
view.

As described in [replication.md](./replication.md), every member update has a
version. Therefore the clients track the versions of their known members. When
they call `Subscribe` they include this set of known members to request any
updates they’ve missed, which the server uses to calculate a diff and send only
the updates the client is missing.

## Reconnect
When the client reconnects, it calls the `Subscribe` RPC again with its most
recent known member versions to get any updates it missed since it last
connected.

# Registration
Clients can register members into the registry.

To enter a member, clients create a bidirectional stream to the `Register` RPC:
* When first creating the stream, the client sends a `REGISTER` update with the
registered members state
* Every `heartbeat_interval` (default 4s) the clients sends a `HEARTBEAT` update
on the stream to notify the server that the member is alive
* Whenever the members state is updated locally, the client sends another
`REGISTER` update with the updated members state
* When the member is unregistered or the client is closed the client sends an
`UNREGISTER` update to remove the member

Each message sent to the server has a `uint64` sequence number that is
incremented with each update. The server responds with an acknowledgement
containing the corresponding sequence number. Therefore if the server doesn’t
acknowledge the request in the `rpc_timeout` (default 5s), the client determines
the connection is closed and reconnects.

This stream is required since the target of the stream is the owner of the
registered member. If the server doesn’t receive heartbeats for the
`heartbeat_timeout`, it will determine the member is down.

## Reconnect
When a client reconnects, it will re-register its members, creating a new stream
for each.

If the client reconnects quickly, it should re-register before the servers
`heartbeat_timeout` to avoid members being marked as `down`.
