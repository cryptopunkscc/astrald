package cc.cryptopunks.astral.mod.content

import android.content.ContentResolver
import android.content.Context
import android.net.Uri
import cc.cryptopunks.astral.mod.Methods
import cc.cryptopunks.astral.client.ext.byte
import cc.cryptopunks.astral.client.ext.string8
import cc.cryptopunks.astral.client.enc.gson.coder

fun Context.contentResolverMethods(): Methods {
    val resolver: ContentResolver = contentResolver
    return mapOf(
        Content.info to {
            val uriString = string8
            val info = resolver.info(uriString)
            write(coder.encode(info).encodeToByteArray())
            byte
        },
        Content.read to {
            val uriString = string8
            val uri = Uri.parse(uriString)
            val input = resolver.openInputStream(uri)!!
            byte = 0
            copyFrom(input)
            byte
        }
    )
}
