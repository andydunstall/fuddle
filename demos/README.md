# Examples

These examples provide services that use Fuddle to manage the cluster, which
can be run using `fuddle demo`.

## [Random Number Service](./random)
The random number service provides a toy example showing a simple use of Fuddle.

The cluster includes two types of node:
* Frontends: Accept client requests and forward to the appropriate backend node,
* Random: Generates a random number

Run the service using `fuddle demo random`.

<p align="center">
  <img src='../assets/images/random-demo.png?raw=true' width='60%'>
</p>
