# Prototari--The Pelotari Network Protocol

The pelotari protocol (prototari) connects machines inside the same private network, and allows them to communicate using UDP messages.
It doesn't guarantee delivery or ordering, as it uses raw UDP packets without acknowledgement mechanisms.
Good for high throughput and cases where missing a few frames isn't an issue (streaming, gaming...).

The protocol has four parts:

1. **Discovery**--Finding available peers inside the private network.
2. **Heartbeat**--Heartbeat messages used to know if registered peers are still up.
3. **Delivery**--Exchange of messages between registered peers.
4. **Disconnect**--When a peer stops to respond to heartbeat messages, it is removed from the list of peers.

The protocol uses two ports:

- `21451`--To receive UDP broadcast messages in the discovery phase.
- `21450`--For all UDP unicast

A maximum number or peers can be specified before starting the program (defaults to `64`).
When the maximum number of peers are registered, the discovery phase refuses to add more peers until a connected peers decides to close their connection.

## 1. Discovery

The discovery process consists of three separate tasks: _broadcasting_, _responding_ and _handshake_.
Broadcasting sends broadcast messages to make the sending machine visible to the rest of the network.
Responding allows other peers that received the broadcast to send a response--a petition to be added as peers.
The handshake happens when the original broadcaster sends a confirmation to the respondent.

A diagram of the discovery process is as follows:

```
+-----------------+     Broadcast (port 21451)      +-----------------+
|   Broadcaster   |-------------------------------->|    Responder    |
+-----------------+       [pelotari?]               +-----------------+
        |                                                  |
        |                                                  |
        |               Response (port 21450)              |
        |<-------------------------------------------------+
        |                 [aupa!]
        |
        V
+-----------------+     Handshake (port 21450)      +-----------------+
|   Broadcaster   |-------------------------------->|    Responder    |
+-----------------+       [dale!]                   +-----------------+
        |                                                   |
        |                                                   |
        | Add to Peer List                 Add to Peer List |
        |                                                   |
        V                                                   V
+-----------------+                                 +-----------------+
|   Broadcaster   |                                 |    Responder    |
|   (Peer List)   |                                 |   (Peer List)   |
+-----------------+                                 +-----------------+
```

Let's take a look at each of the tasks that make up the discovery phase.

### 1a. Broadcasting

Broadcasting happens continuously while the program is running.
It consists on the following steps:

1. The broadcaster sends a UDP broadcast message on port `21451` to the private network's broadcast address (e.g. `192.168.0.255`).
   The message is as follows: `pelotari?`. That is, the string `pelotari?`.
2. Sleep for a fixed amount of time (configurable; defaults to 5 seconds).
3. If the maximum number of peers has been reached, go back to step 2.
4. Go back to step 1.

### 1.b Responding

When a peer receives a broadcast message from another peer, here's what it does:

1. If the broadcaster is already registered as peer, ignore the message and skip the rest of the steps.
2. If the maximum number of peers is already registered, ignore the message and skip the rest of the steps.
3. Send a UDP unicast response to the broadcaster on port `21450` with the message `aupa!`.
4. When the confirmation from the broadcaster arrives, add the broadcaster as peer.
   If the confirmation never arrives, the broadcaster isn't added as peer.

### 1.c Handshake

When the original broadcaster receives a response, here's what it does:

1. If the maximum number of peers was reached, ignore the response and skip the rest of the steps.
2. Add the responding machine as peer.
3. Confirm the registration of the new peer by sending it a unicast UDP message to port `21450`.
   The message should contain the string: `dale!`.

If the responder is added as peer but never received the confirmation message, it will be removed from the peers list by the heartbeat part of the protocol.

## 2. Heartbeat

A heartbeat is a message sent by a computer to those peers from whom it hasn't heard any messages for a configurable amount of time (inactive peer time).
Each peer has a _last seen_ timestamp with the time when the last message from it was received.
These timestamps are used to calculate when a peer hasn't been seen for some time.
Aditionally, each peer has a "missed heartbeats" counter.
When this counter reaches 3 missed heartbeats, the peer is considered disconnected, and hence removed from the registered peers list.

The heartbeat works as follows:

1. The broadcaster loops through its registered peers.
2. For each peer whose "last seen" timestamp is farther in the past than the inactive peer time, send a hearbeat message.
   Heartbeat messages are unicast UDP messages containing the string `hor?` (there?).
3. The broadcaster waits for a maximum amount of configurable time (heartbeat max wait time).
4. If the peer answers with a `hemen nago!` message (I'm here), the "last seen" timestamp is updated, the missed heartbeats counter reset to zero, and the remaining steps skipped.
5. If the broadcaster doesn't receive response during the allowed time window, its missed heartbeats counter is incremented.
6. If the missed heartbeats reaches 3, the peer is removed from the registered peers.

Here's a simplified diagram of the process:

```
+-----------------+     Heartbeat (port 21450)            +-----------------+
|   Broadcaster   |-------------------------------------->|     Peer X      |
+-----------------+               [hor?]                  +-----------------+
        |                                                         |
        | Wait for response                                       |
        |                      Response (port 21450)              |
        |<--------------------------------------------------------+
        |                         [hemen nago!]
        | No Response
        |
        V
+------------------+
|   Broadcaster    |
| Increment missed |
| heartbeats       |
+------------------+
```

## 3. Delivery

A peer can send messages to any of its registered peers.
Messages are "fire and forget," that is, there isn't any acknowledgement mechanism or retries.

A peer should only accept messages from its known peers.
If a message is received and it doesn't match any of the peers it has registered, it should ignore it.
This prevents external agents that haven't gone through the handshake procedure to interfere with the communication.

Messages are sent as UDP packages to the 21450 port.

## 3. Disconnect

When a peer wishes to disconnect, it should send a `agur!` UDP message to each of its registered peers, at port 21450.
No answer is expected from the peers--as soon as all messages have been sent, the peer is free to disconnect.
For those peers that have missed the disconnect message (UDP doesn't guarantee delivery), the heartbeat procedure should handle the clean up.

When a peer receives an `agur!` message from a registered peer, it should immediately remove it from the list.
Messages from unknown peers are ignored.

## Configuration

- **Max. peers**--The maximum number of peers the protocol will attempt to register (defaults to `64`).
- **Broadcast interval**--The amount of time to wait between broadcast messages (defaults to 5 seconds).
- **Inactive peer time**--The amount of time after which, if a peer hasn't sent any message, a heartbeat is sent.
- **Heartbeat max. wait time**--The maximum amount of time the broadcaster waits for the heartbeat response (defaults to 1 second).
