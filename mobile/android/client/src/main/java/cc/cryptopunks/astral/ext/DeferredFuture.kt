package cc.cryptopunks.astral.ext

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CompletableDeferred
import java.util.concurrent.CompletableFuture

fun <T> CompletableFuture<T>.asDeferred(): CompletableDeferred<T> {
    val deferred = CompletableDeferred<T>()
    whenComplete { t, e ->
        when {
            t != null -> deferred.complete(t)
            e != null -> deferred.completeExceptionally(e)
            else -> deferred.cancel(CancellationException("Future completed without result and error"))
        }
    }
    return deferred
}
