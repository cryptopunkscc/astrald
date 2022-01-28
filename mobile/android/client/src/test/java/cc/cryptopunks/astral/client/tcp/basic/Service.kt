package cc.cryptopunks.astral.client.tcp.basic

import cc.cryptopunks.astral.ext.readMessage
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import cc.cryptopunks.astral.gson.GsonCoder
import kotlinx.coroutines.runBlocking

fun main() {
    runBlocking {
        astralTcpNetwork(GsonCoder()).register(
            port = "tcp-test",
        ) {
            while (readMessage(::print));
        }
    }
}
