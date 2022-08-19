package cc.cryptopunks.astral.client.tcp

import cc.cryptopunks.astral.client.enc.encoder
import cc.cryptopunks.astral.client.ext.query
import cc.cryptopunks.astral.client.ext.readMessage
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import org.junit.Assert
import org.junit.Test

class BenchmarkTest {

    @Test
    fun test() {
        val request = "Hello!!".toByteArray()
        val times = 100000
        var received = 0
        val astral = astralTcpNetwork().encoder()
        runBlocking(Dispatchers.IO) {
            // run service
            launch {
                val port = astral.register("tcp-test")
                val conn = port.next()
                val stream = conn().accept()
                repeat(times) {
                    stream.readMessage()
                    received++
                }
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
                    repeat(times) {
                        write(request)
                    }
                }
            }
        }
        Assert.assertEquals(times, received)
    }
}
