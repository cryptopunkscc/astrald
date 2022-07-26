package cc.cryptopunks.astral.wrapper

import android.content.Context
import android.util.Log
import astral.Astral
import cc.cryptopunks.astral.android.resolveMethods
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
import java.io.File
import java.util.concurrent.Executors

const val ASTRAL = "Astral"

private val executor = Executors.newSingleThreadExecutor()

private val scope = CoroutineScope(SupervisorJob() + executor.asCoroutineDispatcher())

private var astralJob: Job = Job().apply { complete() }

private val identity = CompletableDeferred<String>()

private val status = MutableStateFlow(AstralStatus.Stopped)

val astralStatus: StateFlow<AstralStatus> get() = status

enum class AstralStatus { Starting, Started, Stopped }

fun Context.startAstral() {
    if (status.value == AstralStatus.Stopped) {
        status.value = AstralStatus.Starting

        // Start astral daemon
        astralJob = scope.launch {
            val multicastLock = acquireMulticastWakeLock()
            try {
                val dir = File(applicationInfo.dataDir).absolutePath
                val modules = Modules.from(resolveMethods())

                Astral.start(dir, modules)
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                status.value = AstralStatus.Stopped
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

        status.value = AstralStatus.Started
    }
}

fun stopAstral() = runBlocking {
    Astral.stop()
    astralJob.join()
}

suspend fun astralIdentity() = identity.await()
