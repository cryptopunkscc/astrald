package cc.cryptopunks.astral.concurrent

import cc.cryptopunks.astral.enc.EncConnection
import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncPort
import cc.cryptopunks.astral.enc.EncStream
import java.util.concurrent.CompletableFuture
import java.util.concurrent.ExecutorService

interface NetworkExecutor : ExecutorService {
    val network: EncNetwork
    val executor get() = this
}

fun NetworkExecutor.register(
    query: String,
    handle: EncStream.() -> Unit,
): CompletableFuture<CompletableFuture<Unit>> {
    val register = {
        network.register(query)
    }
    val serve = { port: EncPort ->
        val closer = CompletableFuture<Unit>()
        execute {
            try {
                while (true) {
                    val connection = port.next()
                    val handler = handler(connection, handle)
                    execute(handler)
                }
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                closer.cancel(false)
                port.close()
            }
        }
        closer.handle { _, _ ->
            port.close()
        }
    }
    return CompletableFuture
        .supplyAsync(register, executor)
        .thenApply(serve)
}

private fun handler(
    obtain: (() -> EncConnection),
    handle: EncStream.() -> Unit,
) = Runnable {
    try {
        val connection = obtain()
        val stream = connection.accept()
        try {
            stream.handle()
        } catch (e: Throwable) {
            e.printStackTrace()
        } finally {
            stream.close()
        }
    } catch (e: Throwable) {
        e.printStackTrace()
    }
}

fun <T> NetworkExecutor.queryFuture(
    port: String,
    identity: String = "",
    handle: EncStream.() -> T,
): CompletableFuture<T> {
    val completable = CompletableFuture<T>()
    val future = submit {
        try {
            val stream = network.query(port, identity)
            try {
                val result = stream.handle()
                completable.complete(result)
            } catch (e: Throwable) {
                completable.completeExceptionally(e)
            } finally {
                stream.close()
            }
        } catch (e: Throwable) {
            completable.completeExceptionally(e)
        }
    }
    return completable.whenComplete { _, _ ->
        if (!future.isDone) try {
            future.cancel(false)
        } catch (e: Throwable) {
            e.printStackTrace()
        }
    }
}

fun <T : Any> NetworkExecutor.queryFutureResult(
    port: String,
    identity: String = "",
    handle: StreamResult<T>.() -> Unit,
): CompletableFuture<T> {
    return queryFuture(port, identity) {
        StreamResult<T>(this).apply {
            handle()
        }.result
    }
}

class StreamResult<T : Any>(
    stream: EncStream,
) : EncStream by stream {
    lateinit var result: T
}
