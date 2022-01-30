package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncStream
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

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
): T = withContext(Dispatchers.IO) {
    val stream = query(identity, port)
    try {
        stream.handle()
    } catch (e: Throwable) {
        throw e
    } finally {
        stream.close()
    }
}
