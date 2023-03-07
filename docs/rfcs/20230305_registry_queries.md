# RFC: Registry Queries

**Status**: In progress

When users request the state of the registry or subscribe to updates they should
be able to query only the state they are interested in.

This RFC proposes a query format that filters nodes and based on service,
locality and state.

This does not describe for methods for subscribing to node state.

# Requirements

* Must be able to query nodes based on their service, locality and state

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
Filters out nodes whose state doesn't match the filter.

If an entry matches multiple state filter keys, it must match the values for
all of them.

Such as:
* Match nodes whose `status` is `active` and protocol version is 2 or 3: `{ “status”: [“active”], “protocol.version”: [“2”, "3"] }`
* Match nodes whose IP begins with "10.": `{ “addr.ip”: [“10.*”] }`

Similar to the locality, the format of the state keys are user defined though it
is recommended to use a hierarchy where each level is separated by a dot.

# Examples

### Active Order Service Nodes In `us-east-1`
Queries `order` service nodes in `us-east-1` whole `status` is `active` and
`protocol.version` is either 2 or 3.
```
{
    // Storage service only.
    “storage”: {
        // AWS eu-west-2 only.
        locality: [“aws.us-east-1.*”],
        state: {
            “status”: [“active”],
            “protocol.version”: [“2”, "3"],
        }
    }
}
```
