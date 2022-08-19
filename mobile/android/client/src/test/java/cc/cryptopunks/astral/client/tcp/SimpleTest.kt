package cc.cryptopunks.astral.client.tcp

import cc.cryptopunks.astral.client.enc.encoder
import cc.cryptopunks.astral.client.ext.query
import cc.cryptopunks.astral.client.ext.readMessage
import cc.cryptopunks.astral.client.enc.gson.GsonCoder
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
        val astral = astralTcpNetwork().encoder(GsonCoder())
        runBlocking(Dispatchers.IO) {
            // run service
            launch {
                val port = astral.register("tcp-test")
                val conn = port.next()
                val stream = conn().accept()
                val message = stream.readMessage()
                received = message
                stream.close()
                port.close()
            }
            // run client
            launch {
                delay(500)
                astral.query(
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
