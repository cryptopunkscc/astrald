package cc.cryptopunks.astral.service.content

import android.content.ContentResolver
import android.content.Context
import android.net.Uri
import cc.cryptopunks.astral.android.AndroidApi
import cc.cryptopunks.astral.gson.coder

fun Context.contentResolverMethods(
    resolver: ContentResolver = contentResolver,
) = AndroidApi.Methods(
    get = mapOf(
        Content.info to { arg ->
            val uriString = coder.decode(arg, String::class.java)
            val info = resolver.info(uriString)
            coder.encode(info).encodeToByteArray()
        }
    ),
    read = mapOf(
        Content.read to { arg, write ->
            val uriString = coder.decode(arg, String::class.java)
            val uri = Uri.parse(uriString)
            val input = resolver.openInputStream(uri)!!
            input(write)
        }
    )
)

internal object Content {
    data class Info(
        val uri: String,
        val size: Long,
        val mime: String = "",
    )
    const val read = "sys/content"
    const val info = "sys/content/info"
}
