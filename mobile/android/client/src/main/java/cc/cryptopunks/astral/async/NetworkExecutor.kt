package cc.cryptopunks.astral.async

import cc.cryptopunks.astral.concurrent.NetworkExecutor
import cc.cryptopunks.astral.enc.EncStream
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Deferred

suspend fun <T> NetworkExecutor.query(
    port: String,
    identity: String = "",
    handle: EncStream.() -> T,
): T {
    return queryAsync(port, identity, handle).await()
}

fun <T> NetworkExecutor.queryAsync(
    port: String,
    identity: String = "",
    handle: EncStream.() -> T,
): Deferred<T> {
    val deferred = CompletableDeferred<T>()
    val future = submit {
        try {
            val stream = network.query(port, identity)
            try {
                val result = stream.handle()
                deferred.complete(result)
            } catch (e: Throwable) {
                deferred.completeExceptionally(e)
            } finally {
                stream.close()
            }
        } catch (e: Throwable) {
            deferred.completeExceptionally(e)
        }
    }
    deferred.invokeOnCompletion {
        if (!future.isDone) runCatching {
            future.cancel(false)
        }
    }
    return deferred
}
