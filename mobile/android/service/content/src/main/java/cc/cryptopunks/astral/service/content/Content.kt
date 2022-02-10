package cc.cryptopunks.astral.service.content

import android.content.ContentResolver
import android.net.Uri
import android.provider.OpenableColumns
import java.io.InputStream

object Content {
    data class Info(
        val uri: String,
        val size: Long,
        val mime: String = "",
    )

    interface Service {
        fun reader(uri: String): InputStream
        fun info(uri: String): Info
    }
}

class ContentService(
    private val resolver: ContentResolver,
) : Content.Service {

    override fun reader(uri: String): InputStream =
        resolver.openInputStream(Uri.parse(uri))!!

    override fun info(uri: String): Content.Info {
        val parsedUri = Uri.parse(uri)
        return Content.Info(
            uri = uri,
            mime = resolver.getType(parsedUri).orEmpty(),
            size = resolver.run {
                getLengthFromFileDescriptor(parsedUri)
                    ?: getLengthFromContent(parsedUri)
                    ?: -1
            }
        )
    }
}

fun ContentResolver.getLengthFromFileDescriptor(uri: Uri): Long? = try {
    openAssetFileDescriptor(uri, "r")?.length?.takeIf { it != -1L }
} catch (e: Throwable) {
    e.printStackTrace()
    null
}

private fun ContentResolver.getLengthFromContent(uri: Uri): Long? = this
    // if "content://" uri scheme, try contentResolver table
    .takeIf { uri.scheme == ContentResolver.SCHEME_CONTENT }
    ?.query(uri, arrayOf(OpenableColumns.SIZE), null, null, null)?.run {
        try {
            // maybe shouldn't trust ContentResolver for size: https://stackoverflow.com/questions/48302972/content-resolver-returns-wrong-size
            getColumnIndex(OpenableColumns.SIZE)
                .takeIf { it != -1 }
                ?.also { moveToFirst() }
                ?.let { sizeIndex -> getLong(sizeIndex) }
        } catch (e: Throwable) {
            e.printStackTrace()
            close()
            null
        }
    }
