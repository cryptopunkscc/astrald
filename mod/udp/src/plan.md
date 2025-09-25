# Reliable UDP Module: Architectural Brief

## Purpose & Context
The Reliable UDP module provides stream-like semantics over UDP, ensuring ordered and reliable delivery of data. It is designed to integrate seamlessly with the Astral ecosystem, particularly the `exonet` module and node communication. Unlike raw UDP, this module introduces mechanisms for retransmissions, acknowledgments, and a handshake protocol to establish connection state before data exchange.

## Interfaces & Contracts
- **Connection Interface**: Provides ordered, reliable byte streams. Implements `io.ReadWriteCloser`.
- **Listener Behavior**: Accepts incoming connections, demultiplexing based on remote endpoints.
- **Endpoint Handling**: Supports parsing, packing, and unpacking of network addresses.
- **Invariants**:
  - Data is delivered in order.
  - Lost packets are retransmitted.
  - Connections are established via a handshake.

## Handshake Protocol
The handshake follows a three-step process:
1. **SYN**: Initiator sends a SYN packet with an initial sequence number.
2. **SYN|ACK**: Responder replies with a SYN|ACK, acknowledging the initiator's sequence number and providing its own.
3. **ACK**: Initiator acknowledges the responder's sequence number, completing the handshake.

### Sequence Space Rules
- SYN and FIN each consume one sequence number.
- Retransmissions occur if no acknowledgment is received within the retransmission timeout (RTO).
- A connection is established after the ACK is received.

### Timing & Retransmission
- Initial RTO: 500ms (configurable).
- Exponential backoff for retransmissions.
- Maximum retries: 8 (configurable).

## Data Path Overview
- **Segmentation**: Application data is split into packets, each with a sequence number.
- **Ordering**: Out-of-order packets are buffered until missing packets arrive.
- **Acknowledgments**: Cumulative ACKs confirm receipt of all bytes up to a sequence number.
- **Retransmission**: Unacknowledged packets are retransmitted after RTO.
- **Batching**: Multiple packets may be sent together to optimize throughput.

## Concurrency & I/O Model
- **Locking Domains**: Separate locks for send and receive paths.
- **Goroutines**:
  - One for sending data.
  - One for receiving and processing packets.
  - Timers for retransmissions.
- **Shutdown**: Ensures all goroutines exit cleanly, and no resources are leaked.

## Error Model & Shutdown Semantics
- **Errors**: Surface as `net.Error` or module-specific errors.
- **Idempotent Close**: Closing a connection multiple times has no adverse effects.
- **Partial Failures**: Errors during send/receive are propagated to the caller.

## Compatibility & Integration
- **Endpoint Parsing**: Compatible with `exonet` endpoint parsing and unpacking.
- **Lifecycle Alignment**: Designed to align with the lifecycle of other modules like TCP and Tor.
- **Assumptions**: Assumes reliable delivery within the module; does not handle NAT traversal or encryption.

## Security & Future Considerations
- **Stateless Cookie**: Potential for DoS mitigation using stateless cookies during the handshake.
- **PLPMTUD**: Path MTU discovery to avoid fragmentation.
- **Congestion Control**: Future integration with congestion control mechanisms.

## Missing Logics and Potential Issues

### Missing Logics
1. **Connection Handshake Validation**:
   - The handshake process lacks validation for replay attacks or duplicate SYN packets. This could lead to unnecessary resource allocation.

2. **Congestion Control**:
   - The module does not implement congestion control mechanisms, which could lead to network congestion in high-traffic scenarios.

3. **DoS Mitigation**:
   - There is no stateless cookie mechanism during the handshake to prevent denial-of-service (DoS) attacks.

4. **Connection Timeout**:
   - The module does not enforce a timeout for idle connections, which could lead to resource exhaustion.

5. **Error Propagation**:
   - Errors during retransmissions or ACK handling are not consistently propagated to the caller, which could make debugging difficult.

### Potential Performance Issues
1. **Timer Management**:
   - The retransmission timer (`armRTO`) and ACK delay timer (`armAckDelay`) are reset frequently, which could lead to high overhead in timer management.

2. **Lock Contention**:
   - The use of mutexes (`rtoMu`) for timer operations could lead to contention under high concurrency.

3. **Inefficient Buffering**:
   - The `sendQ` buffer in `conn.go` may become a bottleneck if the application writes data faster than the network can transmit.

4. **Packet Parsing Overhead**:
   - The `UnmarshalPacket` function in `recv.go` is called for every incoming packet, which could become a performance bottleneck if the parsing logic is complex.

### Potential Bugs
1. **Timer Race Conditions**:
   - The `armRTO` and `stopRTO` functions do not ensure that the timer callback (`handleRTO`) is not running when the timer is stopped, which could lead to race conditions.

2. **Endpoint Parsing Errors**:
   - The `Dial` function in `dial.go` does not handle errors from `udp.ParseEndpoint`, which could lead to nil pointer dereferences.

3. **Unbounded Retransmissions**:
   - The retransmission logic does not enforce a maximum number of retries, which could lead to infinite retransmissions in case of persistent packet loss.

4. **ACK Timer Reset**:
   - The `armAckDelay` function resets the ACK timer without checking if the timer is already running, which could lead to missed ACKs.

## Implementation Status

### Fully Implemented and Tested
1. **Ring Buffer**:
   - Complete implementation with test coverage for:
     - Blocking write/read operations
     - Buffer closure handling
     - Concurrent access patterns

2. **Packet Serialization**:
   - Full implementation with tests for:
     - Marshal/unmarshal operations
     - Valid packet handling
     - Empty payload cases

3. **Configuration**:
   - Complete implementation with tests covering:
     - Default values
     - Range validation
     - Value normalization

### Fully Implemented but Untested
1. **Data Transmission**:
   - Segmentation and packet sending in `send.go`
   - Retransmission handling in `timers.go`
   - No test coverage for edge cases or error conditions

2. **Data Reception**:
   - Packet processing and buffering in `recv.go`
   - Out-of-order packet handling
   - No tests for complex reassembly scenarios

3. **Server Logic**:
   - Connection management in `server.go`
   - Datagram routing
   - Lacks tests for concurrent connections

### Partially Implemented
1. **Handshake Protocol**:
   - Basic structure defined in `packet.go` (SYN/ACK/FIN flags)
   - Missing implementation in:
     - Connection establishment logic
     - State machine for handshake steps
     - Timeout handling during handshake

2. **Error Handling**:
   - Basic error types defined
   - Inconsistent propagation in retransmission logic
   - Missing comprehensive error recovery

3. **Timer Management**:
   - Basic timer operations implemented
   - Race condition risks identified
   - Missing proper cleanup and synchronization

### Missing Components
1. **Connection State Management**:
   - No explicit connection state machine
   - Missing timeout handling for idle connections
   - No graceful connection termination

2. **Flow Control**:
   - Window size tracking implemented
   - Missing:
     - Congestion control
     - Slow start mechanism
     - Fast retransmit/recovery

3. **Security Features**:
   - No DoS protection
   - Missing replay attack prevention
   - No cookie mechanism for handshake

4. **Testing Infrastructure**:
   - Need integration tests for:
     - Complete connection lifecycle
     - Error scenarios
     - Performance under load
     - Network condition simulation

### Next Steps (Prioritized)
1. **Complete Handshake Implementation**:
   - Implement state transitions
   - Add timeout handling
   - Include sequence number validation

2. **Add Connection Management**:
   - Implement idle connection detection
   - Add connection timeouts
   - Create cleanup mechanisms

3. **Enhance Security**:
   - Add SYN cookie mechanism
   - Implement replay protection
   - Add rate limiting for new connections

4. **Improve Reliability**:
   - Add congestion control
   - Implement proper window management
   - Add fast retransmit/recovery

5. **Complete Test Coverage**:
   - Add integration tests
   - Create network simulation tests
   - Test concurrent connections
