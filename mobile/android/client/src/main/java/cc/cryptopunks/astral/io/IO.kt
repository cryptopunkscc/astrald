package cc.cryptopunks.astral.io

import cc.cryptopunks.astral.net.Stream
import java.io.InputStream
import java.io.OutputStream


fun Stream.reader() = inputStream().reader()
fun Stream.writer() = outputStream().writer()

fun Stream.inputStream(): InputStream = InputStreamWrapper(this::read)
fun Stream.outputStream(): OutputStream = OutputStreamWrapper(this::write)

class InputStreamWrapper(
    private val read: (ByteArray) -> Int,
    private val close: () -> Unit = {},
) : InputStream() {
    private val buff = ByteArray(1)

    override fun read(): Int =
        if (read.invoke(buff) == -1) -1
        else buff[0].toInt()

    override fun read(b: ByteArray): Int = read.invoke(b)

    override fun read(b: ByteArray, off: Int, len: Int): Int =
        if (off == 0 && len == b.size) read.invoke(b)
        else {
            val buff = ByteArray(len)
            val r = read.invoke(buff)
            System.arraycopy(buff, 0, b, off, r)
            r
        }

    override fun close() {
        close.invoke()
    }
}

class OutputStreamWrapper(
    private val write: (ByteArray) -> Int,
    private val close: () -> Unit = {},
) : OutputStream() {

    override fun write(b: Int) {
        val buff = ByteArray(1)
        buff[0] = b.toByte()
        write(buff)
    }

    override fun write(b: ByteArray) {
        write.invoke(b)
    }

    override fun write(b: ByteArray, off: Int, len: Int) {
        write.invoke(b.copyOfRange(off, off + len))
    }

    override fun close() {
        close.invoke()
    }
}
