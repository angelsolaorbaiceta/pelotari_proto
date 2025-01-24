# Pelotari Protocol

A protocol to discover computers connected inside the same private network built on top of UDP.
It allows to find other computers, register them as peers, send and receive messages from them.
The protocol doesn't guarantee delivery of UDP messages; it's a fire a forget protocol.

A pelotari is someone who plays [basque pelota](https://en.wikipedia.org/wiki/Basque_pelota).
The protocol analogy is a pelotari looking for peers to play pelota, asking around in its neighbourhood (the private network).

You can imagine the pelotari (a computer in the local network) going to the village's _frontón_ (the court; the private network) to play _pelota_.
To see who wants to play with him, he screams "pelotari?".
Then, those aroud it who know how to play _pelota_ (other computers inside the private network running the same protocol) answer by screaming back: "aupa!" (hello in basque).
If the original _pelotari_ thinks there's space for more people to play (the maximum number of peers hasn't been reached yet), he allows the respondents to join him by answering with "dale!" (let's do it!).
Then, the _pelotaris_ can play together (talk to each other inside the network) while they listen to some [basque folk music](https://www.youtube.com/watch?v=ONsp-SMT6is).


The protocol is implemented as a Go library: **prototari** (protocol pelotari).


## Usage

TODO: Explain the usage.




