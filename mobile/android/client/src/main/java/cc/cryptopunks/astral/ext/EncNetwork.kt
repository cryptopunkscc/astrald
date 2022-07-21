package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncStream
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch

suspend fun EncNetwork.register(
    port: String,
    handle: suspend EncStream.() -> Unit,
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

suspend fun <T> EncNetwork.query(
    port: String,
    identity: String = "",
    timeout: Long = 5000,
    handle: suspend EncStream.() -> T,
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

suspend fun <T : Any> EncNetwork.queryResult(
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
    stream: EncStream,
) : EncStream by stream {
    lateinit var result: T
}

private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
