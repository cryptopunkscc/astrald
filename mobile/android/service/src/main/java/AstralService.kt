package cc.cryptopunks.astral.service

import android.app.Service
import android.content.Intent
import android.util.Log
import cc.cryptopunks.astral.node.ASTRAL
import cc.cryptopunks.astral.node.AstralStatus
import cc.cryptopunks.astral.node.EmptyConfig
import cc.cryptopunks.astral.node.astralConfig
import cc.cryptopunks.astral.node.astralStatus
import cc.cryptopunks.astral.node.loadAstralConfig
import cc.cryptopunks.astral.node.startAstral
import cc.cryptopunks.astral.node.stopAstral
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class AstralService : Service(), CoroutineScope {

    override val coroutineContext = SupervisorJob() + Dispatchers.IO

    private val tag = ASTRAL + "Service"

    override fun onCreate() {
        startForegroundNotification()
        loadAstralConfig()
        if (astralConfig.value == EmptyConfig) {
            showConfigureAstralNotification()
            stopSelf()
        } else launch {
            Log.d(tag, "Starting astral service")
            startAstral()
        }
    }

    override fun onLowMemory() {
        super.onLowMemory()
        Log.d(tag, "On low memory")
    }

    override fun onTrimMemory(level: Int) {
        super.onTrimMemory(level)
        Log.d(tag, "On trim memory")
    }

    override fun onTaskRemoved(rootIntent: Intent?) {
        super.onTaskRemoved(rootIntent)
        Log.d(tag, "On task removed")
    }

    override fun onDestroy() {
        stopForeground(true)
        if (astralStatus.value != AstralStatus.Stopped) {
            Log.d(tag, "Destroying astral service")
            stopAstral()
        }
        cancel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int = START_STICKY

    override fun onBind(intent: Intent) = null
}
