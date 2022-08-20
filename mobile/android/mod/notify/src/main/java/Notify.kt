package cc.cryptopunks.astral.mod.notification

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat

internal fun Context.notify(manager: NotificationManagerCompat, notifications: List<Notification>) {
    notifications.forEach { notification ->
        manager.notify(notification.id + 100, notification.compat(this))
    }
}

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
    .addAction(action?.android(context))
    .setProgress(progress)
    .build()

private fun Notification.Intent.android(context: Context): PendingIntent {
    val action = action.ifEmpty { Intent.ACTION_VIEW }
    val intent = Intent(action, Uri.parse(uri)).apply {
        flags = Intent.FLAG_ACTIVITY_NEW_TASK
    }
    return when (type) {
        "service" -> PendingIntent.getService(context, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT)
        else -> PendingIntent.getActivity(context, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT)
    }
}

private fun Notification.Action.android(context: Context): NotificationCompat.Action {
    return NotificationCompat.Action(resolveIconId(icon), title, intent?.android(context))
}

private fun NotificationCompat.Builder.setProgress(progress: Notification.Progress?) = apply {
    progress?.run {
        setProgress(max, current, indeterminate)
    }
}

private fun resolveIconId(key: String?): Int = R.drawable.baseline_notification_important_black_24dp
