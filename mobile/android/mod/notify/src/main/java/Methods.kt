package cc.cryptopunks.astral.mod.notification

import android.content.Context
import androidx.core.app.NotificationManagerCompat
import cc.cryptopunks.astral.client.enc.StreamEncoder
import cc.cryptopunks.astral.client.ext.byte
import cc.cryptopunks.astral.client.ext.decodeList
import cc.cryptopunks.astral.client.ext.decodeMessage

fun Context.notificationManagerMethods(): Map<String, StreamEncoder.() -> Unit> {
    val manager: NotificationManagerCompat = NotificationManagerCompat.from(this)
    return mapOf(
        Notification.channel to {
            val channel = decodeMessage<Notification.Channel>()
            manager.create(channel)
            byte = 0
        },
        Notification.notify to {
            while (true) {
                val notifications = decodeList<Notification>()
                notify(manager, notifications)
                byte = 0
            }
        }
    )
}
