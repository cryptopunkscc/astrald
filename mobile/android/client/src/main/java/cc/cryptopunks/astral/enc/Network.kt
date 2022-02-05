package cc.cryptopunks.astral.enc

import cc.cryptopunks.astral.net.Connection
import cc.cryptopunks.astral.net.Network
import cc.cryptopunks.astral.net.Port
import cc.cryptopunks.astral.net.Stream

interface EncNetwork : Network {
    override fun query(port: String, identity: String): EncStream
    override fun register(port: String): EncPort
}

interface EncPort : Port {
    override fun next(): EncConnection
}

interface EncConnection : Connection {
    override fun accept(): EncStream
}

interface EncStream : Stream, Encoder {
    val encoder get() = this
}

interface Encoder {
    fun encode(any: Any): String
    fun <T> decode(string: String, type: Class<T>): T
    fun <T> decodeList(string: String, type: Class<T>): List<T>
    fun <K, V> decodeMap(string: String, key: Class<K>, value: Class<V>): Map<K, V>
}
