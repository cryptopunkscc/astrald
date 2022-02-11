package cc.cryptopunks.astral.service.notification

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.core.app.NotificationChannelCompat
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat

data class Notification(
    val id: Int,
    val channelId: String,
    val contentTitle: String? = null,
    val contentText: String? = null,
    val contentInfo: String? = null,
    val ticker: String? = null,
    val subText: String? = null,
    val smallIcon: String? = null,
    val number: Int = 0,
    val autoCancel: Boolean = false,
    val ongoing: Boolean = false,
    val onlyAlertOnce: Boolean = false,
    val defaults: Int = 0,
    val silent: Boolean = false,
    val priority: Int = NotificationCompat.PRIORITY_DEFAULT,
    val group: String? = null,
    val groupSummary: Boolean = false,
    val progress: Progress? = null,
    val contentIntent: Intent? = null,
) {

    data class Progress(
        val max: Int,
        val current: Int,
        val indeterminate: Boolean,
    )

    data class Intent(
        val action: String,
        val uri: String,
    )

    data class Channel(
        val id: String,
        val importance: Int,
        val name: String,
    )

    interface Service {
        fun create(channel: Channel)
        fun notify(notifications: List<Notification>)
    }
}

class NotificationService(
    private val context: Context,
) : Notification.Service {
    private val manager = NotificationManagerCompat.from(context)

    override fun create(channel: Notification.Channel) {
        manager.createNotificationChannel(channel.compat())
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
    .setDefaults(defaults)
    .setSilent(silent)
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
