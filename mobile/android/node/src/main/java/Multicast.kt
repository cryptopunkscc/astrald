package cc.cryptopunks.astral.node

import android.content.Context
import android.net.wifi.WifiManager
import androidx.core.content.getSystemService

internal fun Context.acquireMulticastWakeLock(): WifiManager.MulticastLock =
    applicationContext.getSystemService<WifiManager>()!!
        .createMulticastLock("multicastLock").apply {
            setReferenceCounted(true)
            acquire()
        }
