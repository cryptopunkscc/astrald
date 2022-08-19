package cc.cryptopunks.astral.node

import cc.cryptopunks.astral.mod.Methods
import cc.cryptopunks.astral.mod.Serve
import cc.cryptopunks.astral.client.enc.Encoder
import cc.cryptopunks.astral.client.enc.StreamEncoder
import cc.cryptopunks.astral.client.enc.gson.GsonCoder
import java.io.InputStream
import java.io.OutputStream

internal class Handlers(
    list: List<Handler>,
) : astral.Handlers {
    private val iterator = list.iterator()
    override fun next(): Handler? = if (iterator.hasNext()) iterator.next() else null

    companion object {
        fun from(
            methods: Methods,
            enc: Encoder = GsonCoder(),
        ) = Handlers(
            list = methods.map { (name, serve) ->
                Handler(name, enc, serve)
            }
        )
    }
}

internal class Handler(
    private val name: String,
    private val enc: Encoder,
    private val serve: Serve,
) : astral.Handler {

    override fun string(): String = name

    override fun serve(conn: astral.Connection) {
        try {
            Connection(enc, conn).serve()
        } catch (e: Throwable) {
            e.printStackTrace()
        } finally {
            conn.close()
        }
    }
}

private class Connection(
    override val encoder: Encoder,
    private val conn: astral.Connection,
) : StreamEncoder,
    Encoder by encoder {

    override val input: InputStream = ConnectionInput(conn)
    override val output: OutputStream = ConnectionOutput(conn)

    override fun read(buffer: ByteArray): Int = input.read(buffer)
    override fun write(buffer: ByteArray): Int {
        conn.write(buffer)
        return buffer.size
    }

    override fun close() {
        conn.close()
    }
}

private class ConnectionOutput(private val conn: astral.Connection) : OutputStream() {

    override fun write(b: Int) {
        conn.write(byteArrayOf(b.toByte()))
    }

    override fun write(b: ByteArray, off: Int, len: Int) {
        if (len == b.size) conn.write(b)
        else conn.write(b.copyOfRange(off, off + len))
    }

    override fun close() = conn.close()
}

private class ConnectionInput(private val conn: astral.Connection) : InputStream() {
    override fun read(b: ByteArray, off: Int, len: Int): Int {
        val r = conn.read(len.toLong())
        System.arraycopy(r, 0, b, off, r.size)
        return r.size
    }

    override fun read(): Int = conn.read(1)[0].toInt()
    override fun close() = conn.close()
}
