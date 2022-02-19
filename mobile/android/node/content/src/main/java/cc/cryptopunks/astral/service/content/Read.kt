package cc.cryptopunks.astral.service.content

import cc.cryptopunks.astral.android.Write
import java.io.InputStream

internal operator fun InputStream.invoke(write: Write) {
    val size = 16 * 1024
    val buffer = ByteArray(size)
    var len: Int
    // For some reason the golang service is not receiving the first byte.
    // To prevent EOF write additional prefix.
    write(ByteArray(1))
    while (true) {
        len = read(buffer)
        when (len) {
            size -> write(buffer)
            -1 -> break
            else -> write(buffer.copyOf(len))
        }
    }
}
