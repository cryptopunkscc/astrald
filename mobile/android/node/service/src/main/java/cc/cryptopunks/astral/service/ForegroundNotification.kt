package cc.cryptopunks.astral.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Build
import androidx.annotation.DrawableRes
import androidx.core.app.NotificationChannelCompat
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.getSystemService
import cc.cryptopunks.astral.service.ui.MainActivity

internal fun Service.startForegroundNotification(@DrawableRes icon: Int) {
    val channelId = createNotificationChannel()

    val pendingIntent: PendingIntent =
        Intent(this, MainActivity::class.java).let { notificationIntent ->
            PendingIntent.getActivity(this, 0, notificationIntent, 0)
        }

    val notification: Notification = NotificationCompat
        .Builder(this, channelId)
        .setSmallIcon(icon)
        .setContentIntent(pendingIntent)
        .setContentTitle("Astral")
        .build()

    // Notification ID cannot be 0.
    startForeground(1, notification)
}


private fun Context.createNotificationChannel(): String = when {
    Build.VERSION.SDK_INT >= Build.VERSION_CODES.O ->
        NotificationChannel(
            "astral",
            "Astral Service",
            NotificationManager.IMPORTANCE_HIGH
        ).apply {
            lightColor = Color.BLUE
//            importance = NotificationManager.IMPORTANCE_NONE
            lockscreenVisibility = Notification.VISIBILITY_PRIVATE
        }.also { channel ->
            requireNotNull(getSystemService<NotificationManager>())
                .createNotificationChannel(channel)
        }.id

    else ->
        NotificationChannelCompat.Builder(
            "astral",
            NotificationManagerCompat.IMPORTANCE_HIGH
        ).apply {
            setLightColor(Color.BLUE)
//            setImportance(NotificationManagerCompat.IMPORTANCE_NONE)
        }.build().also { channel ->
            requireNotNull(getSystemService<NotificationManagerCompat>())
                .createNotificationChannel(channel)
        }.id
}
