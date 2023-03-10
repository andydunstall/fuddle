# Demos

These demos provide services that use Fuddle to manage the cluster, which can
be run using `fuddle demo`.

## [Counter Service](./counter)

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

Run the service using `fuddle demo counter`.

See [`counter/`](./counter) for documentation on the service usage and
architecture.

## [Random Number Service](./random)
The random number service provides a toy example showing a simple use of Fuddle.

The cluster includes two types of node:
* Frontends: Accept client requests and forward to the appropriate backend node,
* Random: Generates a random number

Run the service using `fuddle demo random`.

<p align="center">
  <img src='../assets/images/random-demo.png?raw=true' width='60%'>
</p>
