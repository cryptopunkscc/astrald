# UDP Module

This module implements UDP-based communication for the Astrald platform, enabling fast, connectionless data transfer between nodes.

## Current Proof of Concept (PoC) State
- Basic UDP packet send/receive functionality
- Simple endpoint resolution
- Fragmentation and reassembly of large packets
- Minimal handshake and connection logic
- Configuration via `config.go`

## Key Components
- **Dial:** Initiates outbound UDP connections. Integrates with Astral's context (`astral.Context`), endpoint abstraction (`exonet.Endpoint`), and parses endpoints using UDP utilities. Returns a reliable connection for Astral's exonet system. Similar to TCP's Dial, but uses UDP-specific logic and interacts with Astral's node and identity systems.
- **ResolveEndpoints:** Resolves available UDP endpoints for a node. Uses Astral's identity system (`astral.Identity`) to verify node identity and returns endpoints as `exonet.Endpoint` via Astral's signal utilities (`sig.ArrayToChan`). Connects with Astral's node and endpoint management.
- **Loader:** Loads the UDP module into Astral. Connects with Astral's node (`astral.Node`), asset management (`core/assets`), and logging (`astral/log.Logger`). Loads configuration from assets, parses public endpoints, and registers the module with Astral's core module system for lifecycle management.
- **Unpack:** Handles packet reassembly for Astral's exonet system. Uses UDP endpoint parsing and error handling, returning endpoints for Astral's network abstraction. Connects with Astral's error and endpoint utilities.
- **Server:** Listens for incoming UDP connections. Uses Astral's context and logging, manages connections and server lifecycle, and integrates with Astral's configuration and endpoint parsing. Registers endpoints and connections with Astral's node and router systems.
- 
## Possible Improvements for Production Readiness
- Advanced flow control mechanisms (dynamic window sizing, congestion control)
- Fast retransmissions and selective acknowledgments (SACK)
- Adaptive retransmission timers (RTT estimation)
- Performance optimizations (buffering, batching)
- Support for NAT traversal and hole punching
- Comprehensive test coverage
- Documentation and usage examples

This module is currently a proof of concept and not recommended for production use without further development.
