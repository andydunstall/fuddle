# Members
The registry contains the set of registered members in the cluster.

## Member State
The member state is the state registered by the Fuddle client.

### Attributes
The member attributes describe the member. Attributes are used for filtering
members, such as looking up members in the `orders` service in `us-east-2`, and
observability.

The attributes contain:
* ID (`string`): A unique identifier for the member in the cluster
* Status (`string`): An application defined status for the member (such as
`booting`, `active` or `leaving`)
* Service (`string`): The type of service running on the member (such as
`orders`, `redis` and `storage`)
* Locality
  * Region (`string`)
  * Availability zone (`string`)
* Started (`int64`): The UNIX timestamp in milliseconds that the member started
* Revision (`string`): An identifier for the version of the service running on
the member, such as a Git tag or commit SHA

### Metadata
The member metadata contains a set of arbitrary key-value pairs containing
application defined state.

Metadata is used to share application specific member information with other
members, such as network address, protocol version, member status etc.

## Member Liveness
The members liveness describes whether a member is healthy or not.

Unlike attributes and metadata, which are set by the client that registered the
member, the member's liveness status is set by Fuddle.

The liveness status is either:
* `up`: The member is healthy and sending heartbeats
* `down`: The member is no longer sending heartbeats
* `left`: The member has left the cluster. Left members are kept in the cluster
for a while after leaving to propagate the update and for observability

The members liveness is determined by the failure detector.
