package cc.cryptopunks.astral.service

import android.app.Service
import android.content.Context
import android.content.Intent
import android.util.Log
import cc.cryptopunks.astral.service.ui.cacheLogcat
import cc.cryptopunks.astral.service.ui.clearLogcatCache
import cc.cryptopunks.astral.wrapper.ASTRAL
import cc.cryptopunks.astral.wrapper.startAstral
import cc.cryptopunks.astral.wrapper.stopAstral
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.MainScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class AstralService : Service(), CoroutineScope by MainScope() {

    private val tag = ASTRAL + "Service"

    override fun onCreate() {
        Log.d(tag, "Starting astral service")
        startForegroundNotification(R.mipmap.ic_launcher)
        launch(Dispatchers.IO) { cacheLogcat() }
        startAstral()
    }

    override fun onDestroy() {
        stopForeground(true)
        stopAstral()
        Log.d(tag, "Destroying astral service")
        cancel()
        clearLogcatCache()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int = START_STICKY

    override fun onBind(intent: Intent) = null

    companion object {
        fun intent(context: Context) = Intent(context, AstralService::class.java)
    }
}

