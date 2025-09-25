# Reliable UDP Module

## Overview
This module provides reliable, ordered, and stream-like communication over UDP. It ensures data integrity and delivery through acknowledgments, retransmissions, and a handshake protocol. The module is part of the Astral ecosystem and integrates with `exonet` for endpoint management.

## File & Struct Map
- **config.go**: Defines configuration constants (e.g., retransmission timeouts, buffer sizes).
- **conn.go**: Implements the `Conn` struct, representing a reliable UDP connection. Handles sending, receiving, and retransmissions.
- **endpoint_resolver.go**: Resolves endpoints for the module, integrating with `exonet`.
- **loader.go**: Initializes the module with dependencies.
- **module.go**: Defines the `Module` struct, the entry point for the UDP module.
- **recv.go**: Implements the receive loop, processing incoming packets and acknowledgments.
- **send.go**: Handles segmentation, batching, and sending of data.
- **ring_buffer.go**: Provides a circular buffer for efficient data storage and retrieval.
- **packet.go**: Defines the `Packet` struct and serialization logic.
- **server.go**: Implements the `Server` struct, managing incoming connections and demultiplexing.

## Current Findings & Considerations
- **ACK Handling**: The module uses cumulative acknowledgments to confirm receipt of data up to a specific sequence number. This simplifies state management but requires careful handling of retransmissions to avoid unnecessary duplicates.
- **RTO/Backoff**: Retransmission timeouts are implemented based on RFC 6298, with exponential backoff to handle varying network conditions. This ensures robustness in the face of packet loss.
- **Handshake**: The handshake protocol establishes connection state before data exchange. However, it currently lacks stateless cookie support, which could mitigate DoS attacks by ensuring that resources are only allocated for legitimate connections.

## Datagram Structure
A datagram in this module is represented by the `Packet` struct. It includes the following fields:
- **Seq (uint32)**: Sequence number indicating the first byte of the segment.
- **Ack (uint32)**: Cumulative acknowledgment number, confirming receipt of all bytes up to this sequence number.
- **Flags (uint8)**: Control flags such as SYN, ACK, and FIN.
- **Win (uint16)**: Advertised receive window size in bytes.
- **Len (uint16)**: Length of the payload.
- **Payload ([]byte)**: The actual data being transmitted.

### Fragmentation and Reassembly
- **Fragmentation**: Large application data is segmented into smaller packets, each fitting within the Maximum Segment Size (MSS). This ensures compatibility with network MTU limits and avoids IP-level fragmentation.
- **Reassembly**: On the receiving side, out-of-order packets are buffered and reassembled into the original data stream once all fragments are received.

## Diagrams
### Handshake Protocol
- The handshake establishes a connection between two endpoints before data exchange.
- Steps:
  1. **SYN**: The initiator sends a SYN packet to start the handshake.
  2. **SYN|ACK**: The responder replies with a SYN|ACK packet, acknowledging the initiator's SYN and sending its own sequence number.
  3. **ACK**: The initiator sends an ACK packet to confirm the responder's sequence number.
- Once the ACK is received, the connection is established.

### Data Flow
- **Write Path**:
  - Application data is segmented into smaller packets (segmentation).
  - Packets are serialized and sent over the network (packetization).
- **Network Transmission**:
  - Packets are transmitted over the network, potentially out of order.
- **Read Path**:
  - Received packets are reassembled into the original data stream (reassembly).
  - Cumulative acknowledgments (ACKs) confirm receipt of data up to a specific sequence number.
- Retransmissions occur for lost packets based on retransmission timeouts (RTO).
