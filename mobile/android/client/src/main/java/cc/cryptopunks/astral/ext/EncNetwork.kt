package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncStream
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch
import java.util.concurrent.Executors

suspend fun EncNetwork.register(
    port: String,
    handle: suspend EncStream.() -> Unit,
) = coroutineScope {
    val handler = register(port)
    println("registered: $port")
    while (true) {
        val connection = handler.next()
        println("next connection")
        launch(Dispatchers.IO) {
            val stream = connection.accept()
            println("accepted")
            try {
                stream.handle()
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                stream.close()
                println("closed")
            }
        }
    }
}

suspend fun <T> EncNetwork.query(
    port: String,
    identity: String = "",
    handle: suspend EncStream.() -> T,
): T = scope.async {
    val stream = query(identity, port)
    try {
        stream.handle()
    } catch (e: Throwable) {
        throw e
    } finally {
        stream.close()
    }
}.await()

private val scope = CoroutineScope(
    SupervisorJob() + Executors.newSingleThreadExecutor().asCoroutineDispatcher()
)
