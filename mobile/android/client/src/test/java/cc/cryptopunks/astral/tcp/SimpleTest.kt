package cc.cryptopunks.astral.tcp

import cc.cryptopunks.astral.ext.query
import cc.cryptopunks.astral.ext.readMessage
import cc.cryptopunks.astral.gson.GsonCoder
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import org.junit.Assert
import org.junit.Test

class SimpleTest {

    @Test
    fun test() {
        val request = "Hello!!"
        var received: String? = null
        runBlocking(Dispatchers.IO) {
            // run service
            launch {
                val port = astralTcpNetwork(GsonCoder()).register("tcp-test")
                val conn = port.next()
                val stream = conn.accept()
                val message = stream.readMessage()
                received = message
                stream.close()
                port.close()
            }
            // run client
            launch {
                delay(500)
                astralTcpNetwork(GsonCoder()).query(
                    identity = "",
                    port = "tcp-test",
                ) {
                    write(request.toByteArray())
                }
            }
        }
        Assert.assertEquals(request, received)
    }
}
