package cc.cryptopunks.astral.wrapper

import cc.cryptopunks.astral.android.Methods
import cc.cryptopunks.astral.android.Serve
import cc.cryptopunks.astral.enc.Encoder
import cc.cryptopunks.astral.enc.StreamEncoder
import cc.cryptopunks.astral.gson.GsonCoder
import java.io.InputStream
import java.io.OutputStream

class Modules(
    list: List<Module>,
) : astral.Modules {
    private val iterator = list.iterator()
    override fun next(): Module? = if (iterator.hasNext()) iterator.next() else null

    companion object {
        fun from(
            methods: Methods,
            enc: Encoder = GsonCoder(),
        ) = Modules(
            list = methods.map { (name, serve) ->
                Module(name, enc, serve)
            }
        )
    }
}

class Module(
    private val name: String,
    private val enc: Encoder,
    private val serve: Serve,
) : astral.Module {

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
        val r = conn.read((len - off).toLong())
        System.arraycopy(r, 0, b, off, r.size)
        return r.size
    }

    override fun read(): Int = conn.read(1)[0].toInt()
    override fun close() = conn.close()
}
