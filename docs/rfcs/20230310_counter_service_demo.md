# RFC: Counter Service Demo

**Status**: In progress

This RFC proposes a new demo to show how Fuddle can be used for complex
application specific routing, rather than just simply load balancing among a set
of stateless nodes.

# Requirements
* Must use Fuddle to load balance among a set of nodes using a custom load
balancing policy
* Must use Fuddle for custom request routing using consistent hashing
* Must be easy to start the demo, observe the cluster and interact with the demo

# Counter Service
The counter service provides a WebSocket API where users connect with an ID and
stream a counter of the number of users with that ID.

So when a new user connects with ID `foo`, the counter of users with ID `foo` is
incremented and sent to all users connected with that ID. Similarly when a user
disconnects, the counter is decremented and sent to all connected users with
that ID.

Each user with the same ID must connect to the same counter service node,
therefore to distribute load among multiple nodes, each node is responsible for
a range of IDs using consistent hashing.

## Usage
To start the counter service cluster, users run `fuddle demo counter`. This must
display information including:
* A list of the nodes in the cluster
* A link to the dashboard URL
* Show how Fuddle CLI commands to inspect the cluster
* Show how to add nodes to the cluster (useful for showing rebalancing and
cluster updates)
* Show how to interact with the service

To interact with the service, users run `fuddle demo counter cli <ID>` which
will stream updates to the counter.

To add nodes to the cluster, run `fuddle demo counter –join`, which will create
the same set of nodes but join the existing cluster.

# Architecture
## Services
### Counter
The counter service maintains the counter for each ID. Each node in the service
is responsible for a range of IDs using consistent hashing. Nodes expose a gRPC
interface to stream updates to the counters.

### Frontend
The frontend service is a stateless service that accepts client connections and
routes requests to the time service and counter service.
