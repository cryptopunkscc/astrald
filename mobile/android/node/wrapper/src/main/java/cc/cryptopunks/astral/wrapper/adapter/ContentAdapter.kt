package cc.cryptopunks.astral.wrapper.adapter

import astralmobile.NativeAndroidContentResolver
import astralmobile.Writer
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.service.content.ContentService

internal class ContentResolverAdapter(
    private val service: ContentService,
) : NativeAndroidContentResolver {

    private val coder = GsonCoder()

    override fun info(uri: String): ByteArray {
        val info = service.info(uri)
        return coder.encode(info).toByteArray()
    }

    override fun read(uri: String, writer: Writer) {
        val input = service.reader(uri)
        val size = 16 * 1024
        val buffer = ByteArray(size)
        var len: Int
        // For some reason the golang service is not receiving the first byte.
        // To prevent EOF write additional prefix.
        writer.write(ByteArray(1))
        while(true) {
            len = input.read(buffer)
            when (len) {
                size -> writer.write(buffer)
                -1 -> break
                else -> writer.write(buffer.copyOf(len))
            }
        }
    }
}
