package cc.cryptopunks.astral.service.notification

import androidx.core.app.NotificationCompat

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

    data class Channel(
        val id: String,
        val importance: Int,
        val name: String,
    )

    data class Intent(
        val action: String,
        val uri: String,
    )

    interface Adapter {
        fun create(channel: Channel)
        fun notify(notifications: List<Notification>)
    }
}
