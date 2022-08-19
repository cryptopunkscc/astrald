package cc.cryptopunks.astral.client.tcp

import cc.cryptopunks.astral.client.err.AstralLocalConnectionException
import cc.cryptopunks.astral.client.ext.byte
import cc.cryptopunks.astral.client.ext.identity
import cc.cryptopunks.astral.client.ext.string8
import cc.cryptopunks.astral.client.Connection
import cc.cryptopunks.astral.client.Network
import cc.cryptopunks.astral.client.Port
import cc.cryptopunks.astral.client.Stream
import cc.cryptopunks.astral.client.err.AstralQueryResultError
import java.io.IOException
import java.io.InputStream
import java.io.OutputStream
import java.net.ServerSocket
import java.net.Socket

fun astralTcpNetwork(): Network = TcpNetwork()

object AppHost {
    const val register = "register"
    const val query = "query"
    const val resolve = "resolve"

    const val success = 0x00
    const val errRejected = 0x01
    const val errFailed = 0x02
    const val errTimeout = 0x03
    const val errAlreadyRegistered = 0x04
    const val errUnexpected = 0xff
}

private class TcpNetwork(
    val nodeAddress: String = "127.0.0.1",
    val nodePort: Int = 8625,
    val localAddress: String = nodeAddress,
) : Network {

    val identity by lazy { resolve() }

    override fun register(port: String): TcpPort {
        val serverSocket = ServerSocket(0)
        val tcpStream = TcpStream(astralSocket())
        val localPort = serverSocket.localPort

        tcpStream.runCatching {
            string8 = AppHost.register
            string8 = port
            string8 = "tcp:$localAddress:$localPort"
            checkResultCode()
        }.onFailure {
            throw Network.Exception.Register(port, it)
        }
        return TcpPort(port, serverSocket)
    }

    override fun identity(): String = identity.decodeToString()

    override fun query(port: String, identity: String): TcpStream {
        val stream = TcpStream(astralSocket())
        val id = if (identity.isBlank())
            this.identity else
            identity.encodeToByteArray()

        stream.runCatching {
            stream.string8 = AppHost.query
            stream.identity = id
            stream.string8 = port
            checkResultCode()
        }.onFailure {
            throw Network.Exception.Query(port, identity, it)
        }
        return stream
    }

    private fun resolve(name: String = "localnode"): ByteArray {
        val socket = astralSocket()
        val stream = TcpStream(socket)

        return stream.runCatching {
            string8 = AppHost.resolve
            string8 = name
            checkResultCode()
            identity
        }.onFailure {
            throw Network.Exception.Resolve(name, it)
        }.getOrThrow()
    }

    private fun astralSocket() = try {
        Socket(nodeAddress, nodePort)
    } catch (e: IOException) {
        throw AstralLocalConnectionException(e)
    }
}

private fun Stream.checkResultCode() {
    val errorMessage = errorMessage(byte.toUByte().toInt())
    if (errorMessage != null) throw AstralQueryResultError(errorMessage)
}

private fun errorMessage(code: Int) = when (code) {
    AppHost.success -> null

    // register
    AppHost.errAlreadyRegistered -> "Port already registered"
    AppHost.errFailed -> "Registering port failed"

    // query, resolve
    AppHost.errRejected -> "Query rejected"
    AppHost.errTimeout -> "Query timeout"
    AppHost.errUnexpected -> "Query unexpected error"

    else -> "Unknown error code $code"
}

private class TcpPort(
    val name: String,
    val server: ServerSocket,
) : Port {

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
        val stream = TcpStream(socket)
        return {
            try {
                TcpConnectionRequest(
                    stream = stream,
                    caller = stream.identity,
                    query = stream.string8,
                )
            } catch (e: Throwable) {
                close()
                throw Port.Exception(name, e)
            }
        }
    }
}

private class TcpConnectionRequest(
    private val stream: Stream,
    private val caller: ByteArray,
    private val query: String,
) : Connection {

    override fun accept(): Stream {
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

private class TcpStream(
    private val socket: Socket,
) : Stream {

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
