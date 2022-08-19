package cc.cryptopunks.astral.client.ext

import cc.cryptopunks.astral.client.enc.NetworkEncoder
import cc.cryptopunks.astral.client.enc.StreamEncoder
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch

suspend fun NetworkEncoder.register(
    port: String,
    handle: suspend StreamEncoder.() -> Unit,
) = coroutineScope {
    val handler = register(port)
    println("registered: $port")
    while (true) {
        val connection = handler.next()
        launch(Dispatchers.IO) {
            val stream = connection().accept()
            try {
                stream.handle()
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                stream.close()
            }
        }
    }
}

suspend fun <T> NetworkEncoder.query(
    port: String,
    identity: String = "",
    timeout: Long = 5000,
    handle: suspend StreamEncoder.() -> T,
): T =
    @Suppress("RedundantAsync")
    scope.async {
        val stream = query(port, identity)
        try {
            stream.handle()
        } catch (e: Throwable) {
            throw e
        } finally {
            stream.close()
        }
    }.await()

suspend fun <T : Any> NetworkEncoder.queryResult(
    port: String,
    identity: String = "",
    timeout: Long = 5000,
    handle: suspend StreamResult<T>.() -> Unit,
): T =
    query(port, identity, timeout) {
        StreamResult<T>(this).apply {
            handle()
        }.result
    }

class StreamResult<T : Any>(
    stream: StreamEncoder,
) : StreamEncoder by stream {
    lateinit var result: T
}

private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
