package cc.cryptopunks.astral.wrapper

import cc.cryptopunks.astral.android.AndroidApi
import cc.cryptopunks.astral.android.plus

class AndroidAdapter(
    vararg methods: Methods,
) : AndroidApi(), astral.AndroidApi {

    override val method: Methods = methods.reduce { acc, methods -> acc + methods }

    override fun read(method: String, arg: String, writer: astral.Writer) =
        read(method, arg, writer::write)
}
