package cc.cryptopunks.astral.wrapper

import android.content.Context
import android.util.Log
import astralmobile.Astralmobile
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
private const val ASTRAL_DIR = "astrald"

private val astralScope = CoroutineScope(
    SupervisorJob() + Executors.newSingleThreadExecutor().asCoroutineDispatcher()
)
var astralJob: Job = Job().apply { complete() }
    private set

private val identity = CompletableDeferred<String>()

private val status = MutableStateFlow(AstralStatus.Stopped)

val astralStatus: StateFlow<AstralStatus> get() = status

enum class AstralStatus { Starting, Started, Stopped }

fun Context.startAstral(): Unit =
    if (status.value == AstralStatus.Started) Unit
    else {
        val dir = File(applicationInfo.dataDir)
            .resolve(ASTRAL_DIR)
            .apply { mkdir() }
            .absolutePath

        astralJob = astralScope.launch {
            val multicastLock = acquireMulticastWakeLock()
            try {
                status.value = AstralStatus.Starting
                Astralmobile.start(dir)
            } catch (e: Throwable) {
                e.printStackTrace()
            } finally {
                status.value = AstralStatus.Stopped
                Log.d("AstralNetwork", "releasing multicast")
                multicastLock.release()
            }
        }
        val id = runBlocking {
            flow { while (true) emit(delay(10)) }
                .map { Astralmobile.identity() }
                .first { !it.isNullOrBlank() }
        }
        Log.d("AstralNetwork", id)
        identity.complete(id)
        status.value = AstralStatus.Started
    }

fun stopAstral() = runBlocking {
    Astralmobile.stop()
    astralJob.join()
}

suspend fun astralIdentity() = identity.await()
