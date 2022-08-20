package cc.cryptopunks.astral.mod.notification

import androidx.core.app.NotificationCompat

internal data class Notification(
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
    val action: Action? = null,
) {

    data class Progress(
        val max: Int,
        val current: Int,
        val indeterminate: Boolean,
    )

    data class Intent(
        val type: String,
        val action: String,
        val uri: String,
    )

    data class Channel(
        val id: String,
        val importance: Int,
        val name: String,
    )

    data class Action(
        val icon: String,
        val title: String? = null,
        val intent: Intent? = null,
    )

    companion object {
        const val channel = "android/notify/channel"
        const val notify = "android/notify"
    }
}
