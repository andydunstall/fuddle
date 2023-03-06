# RFC: Registry Queries

**Status**: In progress

When users request the state of the registry or subscribe to updates they should
be able to query only the state they are interested in.

This RFC proposes a query format that filters nodes and node state based on
service, locality and state.

This does not describe for methods for subscribing to node state.

# Requirements

* Must be able to query nodes based on their service
* Must be able to query nodes based on their locality
* Must be able to query nodes based on their state
* Must be able to filter state entries based on their key

# Format
To support the requirements described above, the RFC proposes a query format:
```
{
    "<service filter>": {
        locality: ["<filter>", ...],
        state: {
            "<state filter>": ["<filter>", ...],
            ...
        }
    },
    ...
}
```

Each filter is a string that may include wildcards using the `'*'` character.

## Service Filter
Each node is compared against all service filters. If a node matches any service
filters it must match the query for that service (including `locality` and
`state`).

A node may match multiple services when using wildcard filters.

If a node doesn’t match any services it is discarded.

## Locality Filter
Filters out nodes that don't match one of the locality filters.

Such as:
* Match all nodes: `[“*”]`
* Match all nodes in AWS: `[“aws.*”]`
* Match all nodes in AWS europe: `[“aws.eu-*”]`
* Match all nodes in AWS `us-east-1-b` or GCP `us-east1-c`: `[“aws.us-east-1.us-east-1-b”, “gcp.us-east1.us-east-1-c”]`

As described in [`20230303_registry_api.md`](./20230303_registry_api.md), the
format of the locality is user defined, though it is recommended to use a
hierarchy where each level is separated by a dot. Such as
`<provider>.<region>.<availability zone>`.

## State Filters
State filters can be used to discard nodes and filter what state should be
included in the response.

The keys of the state filter specify what keys should be included, and the
values filter out nodes whose state doesn't match.

If an entry matches multiple state filter keys, it must match the values for
all of them. So its ok to use filters like
`{ “status”: [“active”], “*”: [“*”] }`, which will filter out nodes whose status
is not `active` and include all state for matching nodes.

Such as:
* Include all state: `{ “*”: [“*”] }`
* Include all addresses: `{ “addr.*”: [“*”] }`
* Match nodes whose `status` is `active` or `booting` and include all node state: `{ “status”: [“active”, “booting”], “*”: [“*”] }`
* Match nodes whose `status` is `active` and include the node addresses: `{ “status”: [“active”], “addr.*”: [“*”] }`

Similar to the locality, the format of the state keys are user defined though it
is recommended to use a hierarchy where each level is separated by a dot.

# Examples

### Active Storage Service In `us-east-1`
Say a cluster is partitioning data among a set of storage nodes in the local
region (`eu-west-2`) using consistent hashing. Nodes need to be notified when a
new active storage node joins or leaves the region, therefore can subscribe
using the query:

```
{
    // Storage service only.
    “storage”: {
        // AWS eu-west-2 only.
        locality: [“aws.eu-west-2.*”],
        state: {
            // Filter to only include nodes where status=active.
            “status”: [“active”],
        }
    }
}
```

### Admin and Messaging Service In Europe
Query the created timestamp of admin and messaging service nodes in europe.
```
{
    “admin”: {
        locality: [“aws.eu-*”, “gcp:europe-*”],
        state: {
            “created”: [“*”],
        }
    },
    “messaging”: {
        locality: [“aws.eu-*”, “gcp:europe-*”],
        state: {
            “created”: [“*”],
        }
    }
}
```
