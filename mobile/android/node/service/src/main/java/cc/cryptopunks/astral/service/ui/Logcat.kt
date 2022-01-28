package cc.cryptopunks.astral.service.ui

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.filter
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*

suspend fun cacheLogcat() = logcatFlow().formatAstralLogs().collect(logcatCache::emit)

fun clearLogcatCache() = logcatCache.resetReplayCache()

fun logcatCacheFlow(): Flow<String> = logcatCache

private val logcatCache = MutableSharedFlow<String>(
    replay = 4096,
    onBufferOverflow = BufferOverflow.DROP_OLDEST
)

private fun logcatFlow(): Flow<String> = callbackFlow {
    var process: Process? = null
    launch(Dispatchers.IO) {
        val date = Date(System.currentTimeMillis())

        val format = SimpleDateFormat("MM-dd hh:mm:ss.SSS", Locale.US).format(date)
        process = Runtime.getRuntime().exec(arrayOf("logcat", "-T", format))
        process?.inputStream?.apply {
            try {
                bufferedReader().useLines { lines ->
                    lines.forEach { line ->
                        send(line)
                    }
                }
            } catch (e: Throwable) {
                close()
            }
        }
    }
    awaitClose {
        process?.apply {
            destroy()
        }
    }
}

const val ASTRAL = "Astral"
const val GO_LOG = "GoLog"

private fun Flow<String>.formatAstralLogs(): Flow<String> =
    filter { ASTRAL in it || GO_LOG in it }.map { line ->
        line.split(' ', limit = 5)
            .run { get(1) + " " + get(4) + "\n\n" }
            .split(": ", limit = 2)
            .run { get(0) + ":\n" + get(1) }
    }
