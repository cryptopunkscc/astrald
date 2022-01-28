package cc.cryptopunks.astral.service.notification

import android.content.Context
import cc.cryptopunks.astral.ext.byte
import cc.cryptopunks.astral.ext.decodeL16
import cc.cryptopunks.astral.ext.decodeL16Array
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.gson.GsonCoder
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import cc.cryptopunks.astral.tcp.astralTcpNetwork

private object Port {
    const val NOTIFY = "sys/notify"
    const val CREATE_CHANNEL = "sys/notify/channel"
}


private val astral = astralTcpNetwork(GsonCoder())

suspend fun Context.startNotificationService() {
    try {
        NotificationsAdapter(this).run {
            withContext(Dispatchers.IO) {
                launch { handleCreate() }
                launch { handleNotify() }
            }
        }
    } catch (e: Throwable) {
        e.printStackTrace()
    }
}

private suspend fun Notification.Adapter.handleCreate() {
    println("Starting notification service handle create")
    try {
        astral.register(Port.CREATE_CHANNEL) {
            println("Creating notification channel")
            val channel: Notification.Channel = decodeL16()
            println("Received channel $channel")
            create(channel)
            println("Writing ok")
            byte = 0
            println("Notification channel created")
        }
    } catch (e: Throwable) {
        e.printStackTrace()
    }

}

private suspend fun Notification.Adapter.handleNotify() {
    println("Starting notification service handle notify")
    astral.register(Port.NOTIFY) {
        val notifications = decodeL16Array<Notification>().toList()
        byte = try {
            notify(notifications)
            0
        } catch (e: Throwable) {
            e.printStackTrace()
            1
        }
    }
}
