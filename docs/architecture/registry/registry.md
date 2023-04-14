# Registry
:warning: Note the registry is being replaced with [registry v2](../registryv2/registry.md).

The registry contains the set of registered members.

# Member State
Each registered member includes:
* Status: Whether the member is considered `up` or `down` by Fuddle
* Attributes: A set of attributes used to look up members and for observability,
such as service and locality
* Metadata: A set of arbitrary key-value pairs containing application state

Members are registered by clients. That client is then the authority for that
member. If the client goes down, all members registered by the client are also
considered down. This is because typically a Fuddle client will be registering
its own process, though may also register a 3rd party service running locally,
such as Redis.

## Status
A members status is either `up` or `down`.

The status is decided by Fuddle based on client heartbeats. If Fuddle does not
receive a heartbeat from a client with registered members for the configured
`heartbeat_timeout` (default 20s), all members registered by that client are
considered `down`.

If `down` members don't become healthy again for the configured
`reconnect_timeout` (default 5m), the members are unregistered.

Keeping `down` members in the cluster makes observability easier as you can
inspect the failed members, and allows applications to decide how they handle
`down` members. Such as load balancer may immediately stop trying to route
requests to `down` members, but a database may want to keep them to avoid
excessive rebalancing.

Note applications may choose to add their own application defined member status
using member metadata.

## Attributes
The member attributes describe the member.

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

# Failure Detector
To detect failed members, each client must send a heartbeat every
`heartbeat_interval` (default 5s) to Fuddle.

If Fuddle doesn’t receive a heartbeat for the `heartbeat_timeout` (default 20s),
it will mark all members registered by that client as `down`.

Members stay in the `down` state until either:
* The client reconnects so members move back to the `up` state
* The `reconnect_timeout` (default 5m) is reached and the down members are
unregistered

If a client eventually reconnects after the `reconnect_timeout`, its members
will be re-registered.
