# Pelotari Protocol

A protocol to discover computers connected inside the same private network built on top of UDP.
It allows to find other computers, register them as peers, send and receive messages from them.
The protocol doesn't guarantee delivery of UDP messages; it's a fire a forget protocol.

You can read [the protocol specification here](protocol.md).

A pelotari is someone who plays [basque pelota](https://en.wikipedia.org/wiki/Basque_pelota).
The protocol analogy is a pelotari looking for peers to play pelota, asking around in its neighbourhood (the private network).

You can imagine the pelotari (a computer in the local network) going to the village's _frontón_ (the court; the private network) to play _pelota_.
To see who wants to play with him, he screams "pelotari?".
Then, those aroud it who know how to play _pelota_ (other computers inside the private network running the same protocol) answer by screaming back: "aupa!" (hello in basque).
If the original _pelotari_ thinks there's space for more people to play (the maximum number of peers hasn't been reached yet), he allows the respondents to join him by answering with "dale!" (let's do it!).
Then, the _pelotaris_ can play together (talk to each other inside the network) while they listen to some [basque folk music](https://www.youtube.com/watch?v=ONsp-SMT6is).


The protocol is implemented as a Go library: **prototari** (protocol pelotari).


## Usage

Instantiate a `CommsManager` passing it your desired configuration parameters, or using the default ones:

```go
var (
    config  = prototari.MakeDefaultConfig()
    manager = prototari.MakeManager(config)
)
```

Start the `CommsManager` communications by calling its `Start()`.
This will start sending broadcast messages and automatically registering peers following the handshake procedure.
You can defer stopping the communications, which is done by the `Stop()`  method.
Calling `Stop()` deregisters all peers, but keeps the connections open.
(To close them, you'd call the `Close()` method, as explained below.)

```go
manager.Start()
defer manager.Stop()
```

The peers will be automatically registered and unregistered for you.
If you want to send a message to all peers, use the `CommsManager` `SendMessage()` method:

```go
manager.SendMessage([]byte("My message"))
```
