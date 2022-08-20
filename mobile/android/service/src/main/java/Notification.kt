package cc.cryptopunks.astral.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.graphics.Color
import android.os.Build
import androidx.annotation.DrawableRes
import androidx.core.app.NotificationChannelCompat
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.getSystemService
import cc.cryptopunks.astral.intent.astralActivityIntent

internal fun Service.startForegroundNotification(@DrawableRes icon: Int) {
    val channelId = createNotificationChannel()

    val pendingIntent: PendingIntent = PendingIntent
        .getActivity(this, 0, astralActivityIntent, 0)

    val notification: Notification = NotificationCompat
        .Builder(this, channelId)
        .setSmallIcon(icon)
        .setContentIntent(pendingIntent)
        .setContentTitle("Astral")
        .build()

    // Notification ID cannot be 0.
    startForeground(1, notification)
}


private fun Context.createNotificationChannel(): String {
    val id = "astral"
    val channelName = "Astral Service"
    val importance = NotificationManager.IMPORTANCE_LOW
    val color = Color.BLUE
    val visibility = Notification.VISIBILITY_PRIVATE
    return when {
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.O
        -> NotificationChannel(id, channelName, importance).apply {
            lightColor = color
            lockscreenVisibility = visibility
        }.also { channel ->
            getSystemService<NotificationManager>()
                ?.createNotificationChannel(channel)
                ?: throw Exception("Cannot obtain NotificationManager")
        }.id
        else
        -> NotificationChannelCompat.Builder(id, importance).apply {
            setLightColor(color)
        }.build().also { channel ->
            getSystemService<NotificationManagerCompat>()
                ?.createNotificationChannel(channel)
                ?: throw Exception("Cannot obtain NotificationManagerCompat")
        }.id
    }
}