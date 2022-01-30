package cc.cryptopunks.astral.service.content

import android.content.ContentResolver
import android.net.Uri
import android.provider.OpenableColumns
import java.io.InputStream

class Adapter(
    private val resolver: ContentResolver
) : Content.Resolver {

    override fun reader(uri: String): InputStream =
        resolver.openInputStream(Uri.parse(uri))!!

    override fun info(uri: String): Content.Info {
        val parseUri = Uri.parse(uri)
        return Content.Info(
            uri = uri,
            mime = resolver.getType(parseUri).orEmpty(),
            size = resolver.getContentSize(parseUri)
        )
    }
}

fun ContentResolver.getContentSize(uri: Uri): Long =
    getLengthFromFileDescriptor(uri) ?: getLengthFromContent(uri) ?: -1

private fun ContentResolver.getLengthFromFileDescriptor(uri: Uri): Long? = try {
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
