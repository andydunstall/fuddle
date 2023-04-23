# Registry
> :warning: **In progress**

The registry maintains the set of registered members in the cluster.

## Members
Each member respresents the running instance of a service. Members are
registered by Fuddle clients, plus each Fuddle node registers itself.

Members contain a set of attributes, including the members ID, status, service,
locality and application defined metadata. This member state can be used by
Fuddle clients to filter the set of members of interest, and contains
information needed to interact with the member.

Fuddle also maintains the liveness of each registered member, with status of
either `up`, `down` or `left`.

See [members.md](./members.md) for details.

## Clients
Application nodes interact with Fuddle using on of the Fuddle SDKs, which run
the Fuddle client.

Clients stream the state of the registry to build an eventually consistent
local copy of the registry, which can then be used by the application to query
the registry and subscribe to updates.

Applications also use the Fuddle client to register members into the registry.
Each member has its own connection to Fuddle, and that connection must remain
active and send regular heartbeats for the member to be considered healthy.

See [client.md](./client.md) for details.

## Failure Detector
The failure detector will mark members that miss their heartbeats as `down`
and eventually removes them from the cluster.

See [failure_detector.md](./failure_detector.md) for details.

## Replication
The registry is replicated to every Fuddle node in the cluster. This means:
* Clients can connect to any Fuddle node and stream the full registry
* The cluster can tolerate nodes failing and losing networking with minimal
disruption

All updates are forwarded to every other Fuddle node. To repair any
discrepancies between nodes Fuddle also runs a background repair process where
nodes periodically synchronise their states.

See [replication.md](./replication.md) for details.

## Node Lifecycle
When Fuddle nodes start up, they wait until they have received the registry
state from other replicas before they begin accepting client connections.

When a Fuddle node is shut down, it will stop accepting client connections and
gradually drop all existing client connections. This forces clients to reconnect
to another node to minimise disruption when the Fuddle node shuts down.

## Fault Tolerance
[fault_tolerance.md](./fault_tolerance.md) describes how each supported fault
scenario is handled by Fuddle.

## Metrics
[metrics.md](./metrics.md) describes the available registry metrics.
