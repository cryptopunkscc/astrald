package cc.cryptopunks.astral.tcp

import cc.cryptopunks.astral.enc.EncConnection
import cc.cryptopunks.astral.enc.EncNetwork
import cc.cryptopunks.astral.enc.EncPort
import cc.cryptopunks.astral.enc.EncStream
import cc.cryptopunks.astral.enc.Encoder
import cc.cryptopunks.astral.err.AstralLocalConnectionException
import cc.cryptopunks.astral.ext.decodeL16
import cc.cryptopunks.astral.ext.encodeL16
import cc.cryptopunks.astral.proto.AstralError
import cc.cryptopunks.astral.proto.Request
import cc.cryptopunks.astral.proto.Response
import java.io.Closeable
import java.io.IOException
import java.net.ServerSocket
import java.net.Socket

fun astralTcpNetwork(
    encoder: Encoder,
): EncNetwork =
    TcpNetwork(encoder)

private class TcpNetwork(
    private val encoder: Encoder,
) : EncNetwork {

    override fun register(port: String): EncPort =
        TcpPort(ServerSocket(0), encoder).apply {
            TcpStream(astralSocket(), encoder).apply {
                val request = Request(
                    type = Request.Type.register,
                    port = port,
                    path = ":" + server.localPort
                )
                println("sending request:")
                encodeL16(request)
                val response = decodeL16<Response>()
                println("register response: $response")
                if (response.status != "ok")
                    throw AstralError(response.error)
            }
        }

    // TODO
    override fun identity(): String = ""

    override fun query(port: String, identity: String) =
        TcpStream(astralSocket(), encoder).apply {
            val request = Request(
                type = Request.Type.connect,
                identity = identity,
                port = port,
                path = ":" + socket.localPort,
            )
            encodeL16(request)
            val response = decodeL16<Response>()
            println("connect response: $response")
            if (response.status != "ok")
                throw AstralError(response.error)
        }
}

private class TcpPort(
    val server: ServerSocket,
    private val encoder: Encoder,
) : EncPort {
    override fun close() = server.close()
    override fun next(): TcpConnectionRequest {
        val stream = TcpStream(server.accept(), encoder)
        return try {
            val request = stream.decodeL16<Request>()
            TcpConnectionRequest(
                stream = stream,
                caller = request.identity,
                query = request.port,
            )
        } catch (e: Throwable) {
            println("rejected while reading request.")
            stream.close()
            throw e
        }
    }
}

private class TcpConnectionRequest(
    private val stream: EncStream,
    private val caller: String,
    private val query: String,
) : EncConnection {
    override fun accept() = stream.apply { encodeL16(Response("ok")) }
    override fun caller(): String = caller
    override fun query(): String = query
    override fun reject() = stream.close()
}

private class TcpStream(
    val socket: Socket,
    encoder: Encoder,
) : EncStream, Closeable, Encoder by encoder {
    override val input by lazy { socket.getInputStream().buffered(4096) }
    override val output by lazy { socket.getOutputStream().buffered(4096) }
    override fun close() = try {
        socket.close()
    } catch (e: Throwable) {
        println("Cannot close astral tcp stream: ${e.localizedMessage}")
    }
    override fun read(buffer: ByteArray): Int = input.read(buffer)
    override fun write(buffer: ByteArray): Int = buffer
        .also(output::write)
        .also { output.flush() }
        .size
}

private fun astralSocket() = try {
    Socket("127.0.0.1", 8625)
} catch (e: IOException) {
    throw AstralLocalConnectionException(e)
}
