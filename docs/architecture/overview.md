# Architecture Overview
This document contains a high level overview of the Fuddle architecture.

## Terminology
* Registry: A data structure containing all registered members
* Member: An entry in the registry representing a running instance of a service
* Node: A Fuddle server node that maintains the registry

## Cluster Membership
Fuddle nodes discover one another and detect failure using the SWIM protocol,
which is implemented using the [memberlist](https://pkg.go.dev/github.com/hashicorp/memberlist) library.

## Registry
The registry maintains the set of registered members in the cluster. It is
replicated across all nodes in the cluster so clients can connect to any node
and receive the full registry, and to ensure fault tolerance.

See [registry.md](./registry/registry.md) for details.
