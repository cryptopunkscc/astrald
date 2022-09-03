package cc.cryptopunks.astral.node

import android.content.Context
import android.util.Log
import astral.Astral
import cc.cryptopunks.astral.mod.resolveMethods
import cc.cryptopunks.astral.node.AstralStatus.Started
import cc.cryptopunks.astral.node.AstralStatus.Starting
import cc.cryptopunks.astral.node.AstralStatus.Stopped
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeoutOrNull
import java.io.File
import java.util.concurrent.Executors
import kotlin.time.Duration.Companion.seconds

const val ASTRAL = "Astral"

private val executor = Executors.newSingleThreadExecutor()

private val scope = CoroutineScope(SupervisorJob() + executor.asCoroutineDispatcher())

private var astralJob: Job = Job().apply { complete() }

private val identity = CompletableDeferred<String>()

private val status = MutableStateFlow(Stopped)

val Context.astralDir get() = File(applicationInfo.dataDir)

val File.nodeDir get() = resolve("node").apply { if (!exists()) mkdir() }

var startTime: Long = System.currentTimeMillis(); private set

val astralStatus: StateFlow<AstralStatus> get() = status

enum class AstralStatus { Starting, Started, Stopped }

fun Context.startAstral() {
    if (status.value == Stopped) {
        startTime = System.currentTimeMillis()
        status.value = Starting

        // Start astral daemon
        astralJob = scope.launch {
            val multicastLock = acquireMulticastWakeLock()
            try {
                val dir = astralDir.absolutePath
                val handlers = Handlers.from(resolveMethods())
                val bluetooth = Bluetooth()

                Astral.start(dir, handlers, bluetooth)
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                status.value = Stopped
                Log.d("AstralNetwork", "releasing multicast")
                multicastLock.release()
            }
        }

        // Resolve local node id
        runBlocking {
            val id = flow { while (true) emit(delay(10)) }
                .map { Astral.identity() }
                .first { id -> !id.isNullOrBlank() }
            Log.d("AstralNetwork", id)
            identity.complete(id)
        }

        status.value = Started
    }
}

fun stopAstral() = runBlocking {
    val status = withTimeoutOrNull(5.seconds) {
        status.first { it != Starting }
    }
    if (status != Stopped) {
        Astral.stop()
        astralJob.join()
    }
}

suspend fun astralIdentity() = identity.await()
