package cc.cryptopunks.astral.tcp

import cc.cryptopunks.astral.enc.EncConnection
import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncPort
import cc.cryptopunks.astral.enc.EncStream
import cc.cryptopunks.astral.enc.Encoder
import cc.cryptopunks.astral.err.AstralLocalConnectionException
import cc.cryptopunks.astral.ext.decodeL16
import cc.cryptopunks.astral.ext.encodeL16
import cc.cryptopunks.astral.net.Connection
import cc.cryptopunks.astral.net.Network
import cc.cryptopunks.astral.net.Port
import cc.cryptopunks.astral.proto.AstralError
import cc.cryptopunks.astral.proto.Request
import cc.cryptopunks.astral.proto.Response
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

private class TcpNetwork(
    private val encoder: Encoder,
) : EncNetwork {

    override fun register(port: String): EncPort {
        val serverSocket = ServerSocket(0)
        val tcpStream = TcpStream(astralSocket(), encoder)
        val request = Request(
            type = Request.Type.register,
            port = port,
            path = ":" + serverSocket.localPort
        )
        tcpStream.runCatching {
            encodeL16(request)
            val response = decodeL16<Response>()
            if (response.status != "ok")
                throw AstralError(response.error)
        }.onFailure {
            throw Network.Exception.Register(port, it)
        }
        return TcpPort(port, serverSocket, encoder)
    }

    // TODO
    override fun identity(): String = ""

    override fun query(port: String, identity: String): TcpStream {
        val socket = astralSocket()
        val stream = TcpStream(socket, encoder)
        val request = Request(
            type = Request.Type.connect,
            identity = identity,
            port = port,
            path = ":" + socket.localPort,
        )
        try {
            stream.encodeL16(request)
            val response = stream.decodeL16<Response>()
            if (response.status != "ok")
                throw AstralError(response.error)
        } catch (e: Throwable) {
            throw Network.Exception.Query(port, identity, e)
        }
        return stream
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
            val request = try {
                stream.decodeL16<Request>()
            } catch (e: Throwable) {
                close()
                throw Port.Exception(name, e)
            }
            TcpConnectionRequest(
                stream = stream,
                caller = request.identity,
                query = request.port,
            )
        }
    }
}

private class TcpConnectionRequest(
    private val stream: EncStream,
    private val caller: String,
    private val query: String,
) : EncConnection {

    override fun accept(): EncStream {
        try {
            stream.encodeL16(Response("ok"))
        } catch (e: Throwable) {
            throw Connection.Exception.Accept(this, e)
        }
        return stream
    }

    override fun reject() {
        try {
            stream.close()
        } catch (e: Throwable) {
            throw Connection.Exception.Reject(this, e)
        }
    }

    override fun caller(): String = caller

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
