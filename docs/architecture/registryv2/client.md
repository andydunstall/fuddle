# Client
> :warning: In progress

The Fuddle client is used by applications to subscribe to registry updates and
register local members.

Clients may register zero or more members. Each member has its own stream to
Fuddle which it uses to send updates and regular heartbeats so Fuddle knows it
is still healthy.

Instead of receiving the entire registry, clients can filter what state they
receive by member service and locality. They may also be configured to not
receive any registry state.
