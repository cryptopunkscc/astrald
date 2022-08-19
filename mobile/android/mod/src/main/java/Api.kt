package cc.cryptopunks.astral.mod

import android.content.Context
import cc.cryptopunks.astral.client.enc.StreamEncoder

typealias Serve = StreamEncoder.() -> Unit

typealias Methods = Map<String, Serve>

interface MethodsProvider {
    fun Context.androidMethods(): Methods
}

fun combineMethods(
    vararg methods: Methods,
): Methods =
    methods.fold(emptyMap()) { acc, m -> acc + m }

fun Context.resolveMethods(): Methods =
    (applicationContext as? MethodsProvider)
        ?.run { androidMethods() }
        ?: emptyMap()
