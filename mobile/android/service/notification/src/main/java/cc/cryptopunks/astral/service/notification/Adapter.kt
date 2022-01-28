package cc.cryptopunks.astral.service.notification

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.core.app.NotificationChannelCompat
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat

internal class NotificationsAdapter(
    private val context: Context
) : Notification.Adapter {

    private val manager = NotificationManagerCompat.from(context)

    override fun create(channel: Notification.Channel) {
        NotificationManagerCompat
            .from(context)
            .createNotificationChannel(channel.compat())
    }

    override fun notify(notifications: List<Notification>) {
        notifications.forEach { notification ->
            NotificationManagerCompat.IMPORTANCE_DEFAULT
            manager.notify(notification.id + 100, notification.compat(context))
        }
    }
}

private fun Notification.Channel.compat() = NotificationChannelCompat
    .Builder(id, importance)
    .setName(name)
    .build()

private fun Notification.compat(context: Context) = NotificationCompat
    .Builder(context, channelId)
    .setContentTitle(contentTitle)
    .setContentText(contentText)
    .setContentInfo(contentInfo)
    .setTicker(ticker)
    .setSubText(subText)
    .setNumber(number)
    .setAutoCancel(autoCancel)
    .setSmallIcon(resolveIconId(smallIcon))
    .setOngoing(ongoing)
    .setOnlyAlertOnce(onlyAlertOnce)
    .setGroup(group)
    .setGroupSummary(groupSummary)
    .setPriority(priority)
    .setContentIntent(contentIntent?.android(context))
    .setProgress(progress)
    .build()

private fun Notification.Intent.android(context: Context): PendingIntent {
    val action = action.ifEmpty { Intent.ACTION_VIEW }
    val intent = Intent(action, Uri.parse(uri)).apply {
        flags = Intent.FLAG_ACTIVITY_NEW_TASK
    }
    return PendingIntent.getActivity(context, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT)
}

private fun NotificationCompat.Builder.setProgress(progress: Notification.Progress?) = apply {
    progress?.run {
        setProgress(max, current, indeterminate)
    }
}

private fun resolveIconId(key: String?): Int = R.drawable.baseline_notification_important_black_24dp
