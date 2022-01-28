package cc.cryptopunks.astral.net

/**
 * Network provides access to core network APIs
 */
interface Network {
    fun connect(identity: String, port: String): Stream
    fun register(port: String): Port
    fun identity(): String
}

/**
 * PortHandler is a handler for a locally registered port
 */
interface Port {
    fun close()
    operator fun next(): Connection
}

/**
 * ConnectionRequest represents a connection request sent to a port
 */
interface Connection {
    fun accept(): Stream
    fun caller(): String
    fun query(): String
    fun reject()
}

/**
 * Stream represents a bidirectional stream
 */
interface Stream {
    fun close()
    fun read(buffer: ByteArray): Int
    fun write(buffer: ByteArray): Int
}
