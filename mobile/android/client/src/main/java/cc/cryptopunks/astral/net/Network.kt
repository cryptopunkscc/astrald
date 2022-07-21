package cc.cryptopunks.astral.net

import java.io.Closeable
import java.io.InputStream
import java.io.OutputStream

/**
 * Network provides access to core network APIs
 */
interface Network {
    fun query(port: String, identity: String = ""): Stream
    fun register(port: String): Port
    fun identity(): String

    object Exception {
        class Query(port: String, identity: String, cause: Throwable) :
            kotlin.Exception("$port $identity", cause)

        class Register(port: String, cause: Throwable) :
            kotlin.Exception(port, cause)

        class Resolve(name: String, cause: Throwable) :
            kotlin.Exception(name, cause)
    }
}

/**
 * PortHandler is a handler for a locally registered port
 */
interface Port : Closeable {
    operator fun next(): () -> Connection

    class Exception(
        port: String,
        cause: Throwable,
    ) : kotlin.Exception(port, cause)
}

/**
 * ConnectionRequest represents a connection request sent to a port
 */
interface Connection {
    fun accept(): Stream
    fun caller(): String
    fun query(): String
    fun reject()


    sealed class Exception(
        connection: Connection,
        cause: Throwable,
    ) :
        kotlin.Exception("${connection.query()} ${connection.caller()}", cause) {
        class Accept(connection: Connection, cause: Throwable) : Exception(connection, cause)
        class Reject(connection: Connection, cause: Throwable) : Exception(connection, cause)
    }
}

/**
 * Stream represents a bidirectional stream
 */
interface Stream : Closeable {
    fun read(buffer: ByteArray): Int
    fun write(buffer: ByteArray): Int
    val input: InputStream
    val output: OutputStream
}
