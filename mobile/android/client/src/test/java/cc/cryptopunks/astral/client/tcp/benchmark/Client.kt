package cc.cryptopunks.astral.client.tcp.benchmark

import cc.cryptopunks.astral.ext.query
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import cc.cryptopunks.astral.gson.GsonCoder
import kotlinx.coroutines.runBlocking

fun main() {
    runBlocking {
        astralTcpNetwork(GsonCoder()).query(
            identity = "033c352b239deb28292d48f36e742e8b84ba60ad1abdcc29c669883836203f6b3a",
            port = "tcp-test",
        ) {
            val hello = "hello!!!\n".toByteArray()
            repeat(100000) { write(hello) }
        }
    }
}
