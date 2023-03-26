# Registry API
The registry API is used by clients running on application nodes to subscribe to
registry updates and register members.

Clients subscribe to updates to build a local eventually consistent view of the
cluster that can be queried locally without having to make RPC calls to the
Fuddle server. Note clients can subscribe to updates without registering members.

Clients can also register members with Fuddle.

# Transport
The registry API uses [gRPC](https://grpc.io/) since it provides reasonable
performance, has streaming support and makes it easy to create new SDKs in
different languages.

The service definition is defined in [fuddle-rpc](https://github.com/fuddle-io/fuddle-rpc).

# Member State
Each registered member includes:
* Status: The members status of either `up` or `down`, which is decided by
Fuddle depending on whether the client that registered the members is sending
regular heartbeats
* Attributes: A set of attributes used to filter members, such as by service or
locality
* Metadata: A set of arbitrary key-value pairs containing application state

## Status
The members status is either `up` or `down`. This status is decided by Fuddle
based on client heartbeats. If a client misses heartbeats for the configured
`heartbeat_timeout` all members registered by that client are considered `down`.
If the members don’t become healthy again before the `reconnect_timeout`, they
are unregistered.

Keeping `down` members in the cluster makes observability easier (as you can
inspect the failed members) and allows applications to decide how they handle
`down` members. Such as load balancer may immediately stop trying to route
requests to `down` members, but a database may want to keep them to avoid
excessive rebalancing.

Note applications may choose to add their own application defined member status
using member metadata.

## Attributes
The member attributes are immutable, so once registered cannot be changed.

These are used for filtering members, such as looking up members in the `orders`
service in `us-east-2`. Note metadata can also be used for filtering.

The attributes contain:
* ID (`string`): A unique identifier for the member
* Service (`string`): The type of service running on the member (such as
`orders`, `redis`, `frontend`)
* Locality (`string`): The location of the members. The format of the locality
is user defined though is recommended to be organized into a hierarchy such as
`<provider>.<region>.<zone>` or `<data center>.<rack>` to make it easy to filter
using wildcards
* Created (`int64`): The UNIX timestamp in milliseconds that the member was
created
* Revision (`string`): An identifier for the version of the service running on
the member, such as a Git tag or commit SHA

## Metadata
The member metadata contains a set of arbitrary key-value pairs containing
application defined state. This can be used to include information members need
to know about one another, such as routing information, protocol versions etc.

Unlike attributes the members metadata may be updated, and updates will be sent
to all other members.

# Subscribing
Clients subscribe to updates to build a local eventually consistent view of the
cluster.

When a client connects (including reconnects) they send a digest containing
their known nodes and versions (see versions below).

The Fuddle server will use the digest to work out what updates the client is
missing and send those updates first. Then will forward any further updates to
the registry to the client.

# Versioning
When clients reconnect they need to be able to stream all missed updates. To
avoid having to send the entire registry even if the client was only
disconnected for a couple of seconds, the registry adds versions to each
registered member so the client can detect what updates it missed.

The registry assigns members a version of 1 when they first register, then
increment the version every time the members state is updated.

When a client reconnects, it requests to subscribe again and includes a digest
containing its known member IDs and versions. The server can then use this to
detect:
* Members the client doesn’t know about so should be sent as register updates
* Members the client thinks are in the registry but have left or failed so should
sent as state updates
* Members that have been unregistered so should be sent as unregister updates
* Members the client knows about but have since been updated so should be sent
as state updates

Note when adding clustering, can update versions to also include the owner ID.

# Failure Detector
Clients connect to a Fuddle node.

Each client must send a heartbeat request to the connected node every
`heartbeat_interval` (defaults to 10 seconds).

If the server doesn’t receive a heartbeat for `heartbeat_timeout` (default to 30
seconds), it will mark all members registered by that client as `down`. Members
stay in the `down` state until the `reconnect_timeout` is reached where they are
unregistered.

When a client reconnects, it re-registers all its members. If the members are
still in the registry they are marked as `up` again, otherwise they are
re-registered.

Note when members are re-registered and still in the registry, they must have
the same client ID as they did before. This avoids different clients mistakenly
registering members with the same ID. If a client registers a member with the
same ID as another client Fuddle responds with an `ALREADY_REGISTERED` error.
