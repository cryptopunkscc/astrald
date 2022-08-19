package cc.cryptopunks.astral.intent

import android.content.Intent
import android.net.Uri

fun Uri.intent(
    action: String = Intent.ACTION_VIEW,
    build: Intent.() -> Unit = {},
) = Intent(action, this).apply(build)
