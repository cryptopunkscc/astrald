package cc.cryptopunks.astral.client.tcp.benchmark

import cc.cryptopunks.astral.ext.readMessage
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import cc.cryptopunks.astral.gson.GsonCoder
import kotlinx.coroutines.runBlocking
import java.lang.System.currentTimeMillis


fun main() {
    runBlocking {
        astralTcpNetwork(GsonCoder()).register(
            port = "tcp-test",
        ) {
            val start = currentTimeMillis()
            while (readMessage { });
            val stop = currentTimeMillis()
            println(start)
            println(stop)
            println(stop - start)
        }
    }
}
