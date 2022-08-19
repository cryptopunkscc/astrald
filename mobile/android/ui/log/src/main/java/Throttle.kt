package cc.cryptopunks.astral.ui.log

import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.channelFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex

fun <T> Flow<T>.throttle(
    interval: Long
): Flow<List<T>> = channelFlow {
    val sync = Mutex()
    var lastUpdate = 0L
    var elements: List<T> = emptyList()
    launch {
        collect {
            sync.lock()
            elements = elements + it
            lastUpdate = System.currentTimeMillis()
            sync.unlock()
        }
    }
    while (true) {
        delay(interval)
        if (lastUpdate > interval) {
            sync.lock()
            send(elements)
            elements = emptyList()
            sync.unlock()
        }
    }
}
