## Agglo: A Framework for Lightweight, Flexible Event Stream Processing

Agglo is an experimental event stream processing framework that enables lightweight, reliable and scalable stream processing alongside persistent object storage, key-value storage or stream processing platforms.

Binge (BINary-at-the-edGE) is the main artifact of the framework. As the name implies, it is a single, compiled binary and can run as a stateless daemon, persistent daemon or as a stand-alone command. This allows the same binary to be deployed in edge gateways, as Kubernetes deployments, in load balancers, cloud (lambda) functions, or anywhere else you need to perform stream processing. The deployed artifact is simply a binary and a single JSON config that defines the event processing pipelines and configuration for connecting to external object stores, key-value stores, HTTP/RPC endpoints and stream processing systems.
