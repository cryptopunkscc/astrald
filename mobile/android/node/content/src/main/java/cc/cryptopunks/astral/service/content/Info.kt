package cc.cryptopunks.astral.service.content

import android.content.ContentResolver
import android.net.Uri
import android.provider.OpenableColumns

internal fun ContentResolver.info(
    uriString: String,
    uri: Uri = Uri.parse(uriString),
) = Content.Info(
    uri = uriString,
    mime = getType(uri).orEmpty(),
    size = getLengthFromFileDescriptor(uri)
        ?: getLengthFromContent(uri)
        ?: -1
)

private fun ContentResolver.getLengthFromFileDescriptor(uri: Uri): Long? = try {
    openAssetFileDescriptor(uri, "r")?.length?.takeIf { it != -1L }
} catch (e: Throwable) {
    e.printStackTrace()
    null
}

private fun ContentResolver.getLengthFromContent(uri: Uri): Long? {
    uri.scheme == ContentResolver.SCHEME_CONTENT || return null
    val cursor = query(uri, arrayOf(OpenableColumns.SIZE), null, null, null) ?: return null

    return cursor.runCatching {
        // maybe shouldn't trust ContentResolver for size: https://stackoverflow.com/questions/48302972/content-resolver-returns-wrong-size
        getColumnIndex(OpenableColumns.SIZE)
            .takeIf { it != -1 }
            ?.also { moveToFirst() }
            ?.let { sizeIndex -> getLong(sizeIndex) }
    }.onFailure { throwable ->
        throwable.printStackTrace()
    }.also {
        cursor.close()
    }.getOrNull()
}
