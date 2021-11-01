package cc.cryptopunks.astral.service.ui

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context

internal fun Context.copyToClipboard(text: String) {
    val clipboard = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
    val clip = ClipData.newPlainText("identity", text)
    clipboard.setPrimaryClip(clip)
}
