package cc.cryptopunks.astral.ui.main

import android.content.Context
import android.content.Intent

interface ActionIntentsProvider {
    fun actionIntents(): Map<String, Intent>
}

fun Context.actionIntents(): Map<String, Intent> =
    (applicationContext as? ActionIntentsProvider)
        ?.run { actionIntents() }
        ?: emptyMap()
