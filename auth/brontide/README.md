brontide
==========

I decided to use [lnd][1] implementation of noise handshake protocol, but couldn't import it
raw (since it's tightly coupled with lnd's internals), so I just gutted out the necessary parts
and adjusted the flow for our use case. Definitely needs some work.

[1]: https://github.com/lightningnetwork/lnd