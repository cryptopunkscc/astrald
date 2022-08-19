package cc.cryptopunks.astral.client.enc

import cc.cryptopunks.astral.client.enc.gson.GsonCoder
import cc.cryptopunks.astral.client.Connection
import cc.cryptopunks.astral.client.Network
import cc.cryptopunks.astral.client.Port
import cc.cryptopunks.astral.client.Stream

fun Network.encoder(
    encoder: Encoder = GsonCoder()
) = NetworkEncoder(
    network = this,
    encoder = encoder,
)

data class NetworkEncoder(
    val network: Network,
    val encoder: Encoder,
) : Network by network {

    override fun query(
        port: String,
        identity: String,
    ): StreamEncoder = StreamEncoderImpl(
        encoder = encoder,
        stream = network.query(
            port = port,
            identity = identity,
        )
    )

    override fun register(
        port: String,
    ): PortEncoder = PortEncoder(
        encoder = encoder,
        port = network.register(port)
    )
}

data class PortEncoder(
    val encoder: Encoder,
    val port: Port,
) : Port by port {
    override fun next(): () -> ConnectionEncoder {
        val nextConnection = port.next()
        return {
            ConnectionEncoder(
                encoder = encoder,
                connection = nextConnection()
            )
        }
    }
}

data class ConnectionEncoder(
    val encoder: Encoder,
    val connection: Connection,
) : Connection by connection {
    override fun accept(): StreamEncoder = StreamEncoderImpl(
        encoder = encoder,
        stream = connection.accept()
    )
}

private data class StreamEncoderImpl(
    override val encoder: Encoder,
    val stream: Stream,
) : StreamEncoder,
    Encoder by encoder,
    Stream by stream

interface StreamEncoder : Stream, Encoder {
    val encoder: Encoder get() = this
}

interface Encoder {
    fun encode(any: Any): String
    fun <T> decode(string: String, type: Class<T>): T
    fun <T> decodeList(string: String, type: Class<T>): List<T>
    fun <K, V> decodeMap(string: String, key: Class<K>, value: Class<V>): Map<K, V>
}
