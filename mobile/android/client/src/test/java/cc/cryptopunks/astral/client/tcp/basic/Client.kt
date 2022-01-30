package cc.cryptopunks.astral.client.tcp.basic

import cc.cryptopunks.astral.ext.query
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import kotlinx.coroutines.runBlocking

const val identity = ""

fun main() {
    runBlocking {
        astralTcpNetwork(GsonCoder()).query(
            identity = identity,
            port = "tcp-test",
        ) {
            write("hello!!!\n".toByteArray())
        }
    }
}
