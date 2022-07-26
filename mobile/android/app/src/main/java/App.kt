package cc.cryptopunks.astral

import android.app.Application
import android.content.Context
import cc.cryptopunks.astral.android.MethodsProvider
import cc.cryptopunks.astral.android.combineMethods

class App :
    Application(),
    MethodsProvider {

    override fun Context.androidMethods() = combineMethods(
        // add android methods here
    )
}
