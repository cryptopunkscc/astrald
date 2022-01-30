package cc.cryptopunks.astral.enc

import cc.cryptopunks.astral.net.Connection
import cc.cryptopunks.astral.net.Network
import cc.cryptopunks.astral.net.Port
import cc.cryptopunks.astral.net.Stream

interface EncNetwork : Network {
    override fun query(identity: String, port: String): EncStream
    override fun register(port: String): EncPort
}

interface EncPort : Port {
    override fun next(): EncConnection
}

interface EncConnection : Connection {
    override fun accept(): EncStream
}

class EncStream(
    stream: Stream,
    val encoder: Encoder,
) : Stream by stream

interface Encoder {
    fun encode(any: Any): String
    fun <T> decode(string: String, type: Class<T>): T
    fun <T> decodeArray(string: String, type: Class<T>): Array<T>
    fun <K, V> decodeMap(string: String, key: Class<K>, value: Class<V>): Map<K, V>
}
