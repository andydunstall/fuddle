# Counter Service

> :warning: **Counter service is still in development**

The counter service provides a simple service that provides a WebSocket endpoint
users connect to and specify an ID. The service then streams the number of users
with that same ID.

So when a new user connects with ID `foo`, the counter of users with ID `foo` is
incremented and sent to all users connected with that ID. Similarly when a user
disconnects, the counter is decremented and sent to all connected users with
that ID.

Each user with the same ID must connect to the same counter service node,
therefore to distribute load among multiple nodes, each node is responsible for
a range of IDs using consistent hashing.

Although this is a simple service, it shows how Fuddle can be used to:
* Observe the nodes in the cluster,
* Discovery the nodes in the cluster and their services,
* Route requests to different nodes using application specific routing,
including using consistent hashing, and load balancing with a custom policy

## Usage

If you haven't already, download the latest `fuddle` binary for you're platform
from [releases](https://github.com/andydunstall/fuddle/releases).

Then start the demo running `fuddle demo counter`. This will spin up the cluster
which can be observed via the Fuddle dashboard at
[`http://127.0.0.1:8221`](http://127.0.0.1:8221).

<p align="center">
  <img src='../../assets/images/counter-service-dashboard.png?raw=true' width='80%'>
</p>
