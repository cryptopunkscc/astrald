package cc.cryptopunks.astral.tcp

import cc.cryptopunks.astral.enc.*
import cc.cryptopunks.astral.err.AstralLocalConnectionException
import cc.cryptopunks.astral.ext.byte
import cc.cryptopunks.astral.ext.identity
import cc.cryptopunks.astral.ext.stringL8
import cc.cryptopunks.astral.net.Connection
import cc.cryptopunks.astral.net.Network
import cc.cryptopunks.astral.net.Port
import cc.cryptopunks.astral.proto.AstralError
import cc.cryptopunks.astral.proto.Request
import java.io.Closeable
import java.io.IOException
import java.io.InputStream
import java.io.OutputStream
import java.net.ServerSocket
import java.net.Socket

fun astralTcpNetwork(
    encoder: Encoder,
): EncNetwork =
    TcpNetwork(encoder)

object AppHost {
    const val success = 0x00
    const val errRejected = 0x01
    const val errFailed = 0x02
    const val errTimeout = 0x03
    const val errAlreadyRegistered = 0x04
    const val errUnexpected = 0xff
}

private class TcpNetwork(
    private val encoder: Encoder,
) : EncNetwork {

    val identity by lazy { resolve() }

    override fun register(port: String): EncPort {
        val serverSocket = ServerSocket(0)
        val tcpStream = TcpStream(astralSocket(), encoder)
        tcpStream.runCatching {
            stringL8 = Request.Type.register.name
            stringL8 = port
            stringL8 = "tcp:127.0.0.1:${serverSocket.localPort}"

            when (val errorCode = byte.toInt()) {
                AppHost.success -> Unit
                AppHost.errAlreadyRegistered -> throw AstralError("port already registered")
                AppHost.errFailed -> throw AstralError("registering port failed")
                else -> throw AstralError("Unknown error code $errorCode")
            }
        }.onFailure {
            throw Network.Exception.Register(port, it)
        }
        return TcpPort(port, serverSocket, encoder)
    }

    override fun identity(): String = identity.decodeToString()

    override fun query(port: String, identity: String): TcpStream {
        val socket = astralSocket()
        val stream = TcpStream(socket, encoder)
        val id = if (identity.isBlank())
            this.identity else
            identity.encodeToByteArray()

        stream.runCatching {
            stream.stringL8 = Request.Type.query.name
            stream.identity = id
            stream.stringL8 = port

            when (val errorCode = byte.toInt()) {
                AppHost.success -> Unit
                AppHost.errRejected -> throw AstralError("Query rejected")
                AppHost.errTimeout -> throw AstralError("Query timeout")
                AppHost.errUnexpected -> throw AstralError("Query unexpected error")
                else -> throw AstralError("Unknown error code $errorCode")
            }
        }.onFailure {
            throw Network.Exception.Query(port, identity, it)
        }
        return stream
    }

    private fun resolve(name: String = "localnode"): ByteArray {
        val socket = astralSocket()
        val stream = TcpStream(socket, encoder)

        return stream.runCatching {
            stringL8 = Request.Type.resolve.name
            stringL8 = name

            when (val errorCode = byte.toInt()) {
                AppHost.success -> identity
                AppHost.errRejected -> throw AstralError("Query rejected")
                AppHost.errTimeout -> throw AstralError("Query timeout")
                AppHost.errUnexpected -> throw AstralError("Query unexpected error")
                else -> throw AstralError("Unknown error code $errorCode")
            }
        }.onFailure {
            throw Network.Exception.Resolve(name, it)
        }.getOrThrow()
    }
}

private class TcpPort(
    val name: String,
    val server: ServerSocket,
    private val encoder: Encoder,
) : EncPort {
    override fun close() {
        try {
            server.close()
        } catch (e: Throwable) {
            println("Cannot close astral tcp port")
            e.printStackTrace()
        }
    }

    override fun next(): () -> TcpConnectionRequest {
        val socket = server.accept()
        val stream = TcpStream(socket, encoder)
        return {
            try {
                TcpConnectionRequest(
                    stream = stream,
                    caller = stream.identity,
                    query = stream.stringL8,
                )
            } catch (e: Throwable) {
                close()
                throw Port.Exception(name, e)
            }
        }
    }
}

private class TcpConnectionRequest(
    private val stream: EncStream,
    private val caller: ByteArray,
    private val query: String,
) : EncConnection {

    override fun accept(): EncStream {
        try {
            stream.byte = 0
        } catch (e: Throwable) {
            throw Connection.Exception.Accept(this, e)
        }
        return stream
    }

    override fun reject() {
        try {
            stream.byte = 1
            stream.close()
        } catch (e: Throwable) {
            throw Connection.Exception.Reject(this, e)
        }
    }

    override fun caller(): String = caller.decodeToString()

    override fun query(): String = query

}

class TcpStream(
    private val socket: Socket,
    encoder: Encoder,
) : EncStream, Closeable, Encoder by encoder {

    override val input: InputStream by lazy { socket.getInputStream() }
    override val output: OutputStream by lazy { socket.getOutputStream() }

    override fun read(buffer: ByteArray): Int =
        input.read(buffer)

    override fun write(buffer: ByteArray): Int {
        output.write(buffer)
        output.flush()
        return buffer.size
    }

    override fun close() {
        try {
            socket.close()
        } catch (e: Throwable) {
            println("Cannot close astral tcp stream")
            e.printStackTrace()
        }
    }
}

private fun astralSocket() = try {
    Socket("127.0.0.1", 8625)
} catch (e: IOException) {
    throw AstralLocalConnectionException(e)
}
