package cc.cryptopunks.astral.client.ext

import cc.cryptopunks.astral.client.byte
import cc.cryptopunks.astral.client.bytes
import cc.cryptopunks.astral.client.int
import cc.cryptopunks.astral.client.long
import cc.cryptopunks.astral.client.short
import cc.cryptopunks.astral.client.Stream
import java.io.EOFException

// =========================== Read ===========================

fun Stream.read(
    size: Int,
) = ByteArray(size)
    .also { buff ->
        val len = read(buff)
        if (len == -1) throw EOFException("EOF")
        check(len == size) { "Expected $size bytes but was $len" }
    }

fun Stream.readMessage(): String? {
    val result = StringBuilder()
    val buffer = ByteArray(4096)
    var len: Int
    do {
        len = read(buffer)
        if (len > 0) result.append(String(buffer.copyOf(len)))
    } while (len == buffer.size)
    return when {
        len == -1 && result.isEmpty() -> null
        else -> result.toString().takeIf { it != "null" }
    }
}

// =========================== Write ===========================

fun Stream.write(
    bytes: ByteArray,
    size: ByteArray,
) {
    write(size)
    write(bytes)
}

fun Stream.write(
    bytes: ByteArray,
    formatSize: Int.() -> ByteArray,
) = write(
    bytes = bytes.size.formatSize(),
    size = bytes
)

var Stream.byte: Byte
    get() = read(Byte.SIZE_BYTES).byte
    set(value) {
        write(value.bytes)
    }

var Stream.short: Short
    get() = read(Short.SIZE_BYTES).short
    set(value) {
        write(value.bytes)
    }

var Stream.int: Int
    get() = read(Int.SIZE_BYTES).int
    set(value) {
        write(value.bytes)
    }

var Stream.long: Long
    get() = read(Long.SIZE_BYTES).long
    set(value) {
        write(value.bytes)
    }

var Stream.bytes8: ByteArray
    get() = read(read(Byte.SIZE_BYTES).byte.toUByte().toInt())
    set(bytes) {
        write(bytes, bytes.size.toByte().bytes)
    }

var Stream.bytes16: ByteArray
    get() = read(read(Short.SIZE_BYTES).short.toUShort().toInt())
    set(bytes) {
        write(bytes, bytes.size.toShort().bytes)
    }

var Stream.bytes32: ByteArray
    get() = read(read(Int.SIZE_BYTES).int)
    set(bytes) {
        write(bytes, bytes.size.bytes)
    }

var Stream.string8: String
    get() = bytes8.decodeToString()
    set(string) {
        bytes8 = string.encodeToByteArray()
    }

var Stream.string16: String
    get() = bytes16.decodeToString()
    set(string) {
        bytes16 = string.encodeToByteArray()
    }

var Stream.string32: String
    get() = bytes32.decodeToString()
    set(string) {
        bytes32 = string.encodeToByteArray()
    }

var Stream.identity: ByteArray
    get() {
        val id = read(33)
        return when {
            id.all { it == zero } -> ByteArray(0)
            else -> id
        }
    }
    set(value) {
        val id = when {
            value.isEmpty() || value.contentEquals(localnode) -> ByteArray(33)
            value.size == 33 -> value
            else -> throw IllegalArgumentException("Invalid identity size ${value.size}, have to be 33 or 0")
        }
        write(id)
    }

private val zero = 0.toByte()
private val localnode = "localnode".toByteArray()
