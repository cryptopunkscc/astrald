package cc.cryptopunks.astral.mod.notification

import androidx.core.app.NotificationChannelCompat
import androidx.core.app.NotificationManagerCompat

internal fun NotificationManagerCompat.create(channel: Notification.Channel) {
    val compat = channel.compat()
    createNotificationChannel(compat)
}

private fun Notification.Channel.compat() = NotificationChannelCompat
    .Builder(id, importance)
    .setName(name)
    .build()
