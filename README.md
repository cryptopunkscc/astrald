astrald
=======

This is a proof-of-concept implementation of astral - a logical communication network. Like most proof-of-concepts,
this code is hightly experimental and will break and crash and might not work for you at all. At this stage I publish
this for developers to get early feedback on the ideas expressed by the code.

### A Logical Communication Network?

We live in a world with many communication networks. Most of us use their LANs at home, PSTN for voice calls and,
obviosuly, the internet on a daily basis. Each of these networks solve 3 critical issues:

* Addressing - a way to refer to a node (phone number, IP, MAC, onion address)
* Routing - finding a path between two nodes
* Transfer - delivering packets of data from one node to another

They do not, however, provide a solution to problems like authentication or encryption. This is left for the
applications to implement (and for good reasons). Consequently, developers spend countless hours implementing secure
connectivity for their apps. Since security is very difficult, developers tend to choose centralized solutions to their
problems, at the price of privacy and freedom of expression. These solutions tend to be havily tied to the underlying
physical network, making them vulnerable to internet blackouts, which might be a problem in areas with poor connectivity
or oppresive governments.

I believe we're missing a network layer. Not a physical one, but one that would be an abstraction over all other
physical networks and give developers freedom to build secure apps that automatically adapt to any connectivity
conditions. Apps that will keep working as long as any two devices can establish a bidirectional bitstream over any
transport. Apps that maximize privacy and freedom of their users. Apps that cannot be disappeared becuase they go
against interests of certain powerful groups.

I propose astral - a logical communication network. Astral API is very similar to basic TCP sockets, but instead of
IP addresses it uses cryptographic keys and instead of uint16 ports, it uses strings. An example echo server looks
like this:

```go
func main() {
    port, err := astral.Listen("echo")
    if err != nil {
        panic(err)
    }

	for req := range port.Requests() {
		fmt.Println("Connection from", req.Caller())
		conn, err := req.Accept()
		if err != nil {
			continue
		}

		go io.Copy(conn, conn)
	}
}
```

The service can check the caller's identity before making the decision whether to accept the request or not. If it
decides to reject it, the caller cannot tell whether the request was rejected, or the port was closed. On accept,
the service gets a bidirectional, encrypted and authenticated bytestream with the caller. 

Here is an example client app:

```go
func main() {
	echo, err := astral.Dial("021b413c350f9da8fe23ee6d145a1351ad9b96ff4acd46fdbe3e53a8b6400c331b", "echo")
	if err != nil {
		log.Panic(err)
	}

	var msg = []byte("hello, astral")
	echo.Write(msg)
	echo.Read(msg)
	echo.Close()
}
```

On success, the client gets a bidirectional, encrypted and authenticated bytestream with the requested node.

Astral node is responsible for establishing which physical network to use and for figuring out the network-specific
address of the node. It employs (or rather will employ) various strategies to meet that goal, but the nature of
distributed systems makes it impossible to do this with 100% efficacy. However, since this problem is invisible to
application developers, all astral apps will reap the benefits of future improvements to this issue.

In the future, once two nodes establish a link with each other, they will exchange all their physical network locations,
so that they can keep an up-to-date record on how to reach each other. For example, if you introduce two nodes to each
other on your local network at home and one of them leaves the LAN, they will already know how to try to reconnect via
the internet, tor, bluetooth, etc. All completely transparent to the app developers.

### Standard Services

I designed the astral protocol to be as lightweight as possible, so that the core API can be implemented quickly
on almost any device. It's intentionally bare bones and strictly scoped. I doubt this is the final form of the
protocol, but the scope is pretty final - simple, secure, resilient connectivity. While I think this is a step in the
right direction, decentralized application developers face many more problems that make them less efficient. I plan
to develop a set of standard services that applications can optionally use to accelerate the development. Such services
would include:

* Service discovery
* Pubsub
* Data exchange
* Node groups 
* Astral-native routing
* etc

The ultimate goal is to make it stupid easy to build complex decentralized apps with strong security, high network
adaptability and seamless scalability of peer-to-peer networks.

### Contact

You can reach me via email or XMPP: arashi@cryptopunks.cc