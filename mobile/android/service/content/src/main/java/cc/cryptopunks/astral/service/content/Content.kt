package cc.cryptopunks.astral.service.content

import java.io.InputStream

object Content {
    data class Info(
        val uri: String,
        val size: Long,
        val mime: String = "",
    )

    interface Resolver {
        fun reader(uri: String): InputStream
        fun info(uri: String): Info
    }
}
