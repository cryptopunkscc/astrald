package cc.cryptopunks.astral.mod.content

import cc.cryptopunks.astral.client.Stream
import java.io.InputStream

internal fun Stream.copyFrom(input: InputStream) {
    val size = 16 * 1024
    val buffer = ByteArray(size)
    var len: Int
    while (true) {
        len = input.read(buffer)
        when (len) {
            size -> write(buffer)
            -1 -> break
            else -> write(buffer.copyOf(len))
        }
    }
}
