package cc.cryptopunks.astral.ui.log

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.flow.filter
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*

internal fun logcatFlow(since: Date): Flow<String> = callbackFlow {
    var process: Process? = null
    launch(Dispatchers.IO) {
        val sinceDate = SimpleDateFormat("MM-dd HH:mm:ss.SSS", Locale.getDefault()).format(since)
        process = Runtime.getRuntime().exec(arrayOf("logcat", "-T", sinceDate))
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

internal fun clearLogcatProcess() {
    try {
        Runtime.getRuntime().exec(arrayOf("logcat", "-c"))
    } catch (e: Throwable) {
        println("Cannot clear logcat")
        e.printStackTrace()
    }
}

internal fun Flow<String>.formatAstralLogs(): Flow<String> =
    filter { line ->
        filterTags.any { it in line }
    }.map { line ->
        line.split(' ', limit = 5)
            .run { get(1) + " " + get(4) }
            .split(": ", limit = 2)
            .run { get(0) + ":\n" + get(1) }
    }

private val filterTags = listOf(
    "Astral",
    "GoLog",
)
