package cc.cryptopunks.astral.service.notification

import android.content.Context
import cc.cryptopunks.astral.ext.byte
import cc.cryptopunks.astral.ext.decodeL16
import cc.cryptopunks.astral.ext.decodeL16List
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import kotlinx.coroutines.launch
import kotlinx.coroutines.supervisorScope

private object Port {
    const val NOTIFY = "sys/notify"
    const val CREATE_CHANNEL = "sys/notify/channel"
}


private val astral = astralTcpNetwork(GsonCoder())

suspend fun Context.startNotificationService() {
    try {
        Adapter(this).run {
            supervisorScope {
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
    try {
        astral.register(Port.NOTIFY) {
            val notifications = decodeL16List<Notification>()
            val result: Byte = try {
                notify(notifications)
                0
            } catch (e: Throwable) {
                println("Cannot display notification")
                e.printStackTrace()
                1
            }
            try {
                byte = result
            } catch (e: Throwable) {
                println("Cannot send notification result")
                e.printStackTrace()
            }
        }
    } catch (e: Throwable) {
        e.printStackTrace()
    }
}
