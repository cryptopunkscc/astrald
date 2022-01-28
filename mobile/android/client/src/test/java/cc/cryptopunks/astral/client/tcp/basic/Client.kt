package cc.cryptopunks.astral.client.tcp.basic

import cc.cryptopunks.astral.ext.connect
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import kotlinx.coroutines.runBlocking

const val identity = ""

fun main() {
    runBlocking {
        astralTcpNetwork(GsonCoder()).connect(
            identity = identity,
            port = "tcp-test",
        ) {
            write("hello!!!\n".toByteArray())
        }
    }
}
