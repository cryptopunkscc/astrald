package cc.cryptopunks.astral.wrapper.adapter

import astralmobile.NativeAndroidNotify
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.service.notification.Notification
import cc.cryptopunks.astral.service.notification.NotificationService

internal class NotifyAdapter(
    private val service: NotificationService,
) : NativeAndroidNotify {

    private val coder = GsonCoder()

    override fun create(bytes: ByteArray) {
        val string = bytes.decodeToString()
        val channel = coder.decode(string, Notification.Channel::class.java)
        service.create(channel)
    }

    override fun notify(bytes: ByteArray) {
        val string = bytes.decodeToString()
        val notificationList = coder.decodeList(string, Notification::class.java)
        service.notify(notificationList)
    }
}
