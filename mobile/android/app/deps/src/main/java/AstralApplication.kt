package cc.cryptopunks.astral.app

import android.content.Context
import android.content.Intent
import cc.cryptopunks.astral.mod.MethodsProvider
import cc.cryptopunks.astral.mod.combineMethods
import cc.cryptopunks.astral.ui.main.ActionIntentsProvider

interface AstralApplication :
    MethodsProvider,
    ActionIntentsProvider {

    override fun Context.androidMethods() = combineMethods(
        // TODO add methods
    )

    override fun actionIntents(): Map<String, Intent> = mapOf(
        // TODO add intents
    )
}
