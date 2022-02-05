package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.binary.byte
import cc.cryptopunks.astral.binary.bytes
import cc.cryptopunks.astral.binary.int
import cc.cryptopunks.astral.binary.long
import cc.cryptopunks.astral.binary.short
import cc.cryptopunks.astral.net.Stream
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

fun Stream.read(
    size: Stream.() -> Number,
): ByteArray =
    read(size().toInt())

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
        result.contains("null") -> null
        else -> result.toString()
    }
}

fun Stream.readMessage(handle: (String) -> Unit): Boolean =
    readMessage()?.let(handle) != null


fun Stream.readL64Bytes(): ByteArray =
    read(long.toInt())

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

var Stream.bytesL8: ByteArray
    get() = read(read(Byte.SIZE_BYTES).byte.toUByte().toInt())
    set(bytes) {
        write(bytes, bytes.size.toByte().bytes)
    }

var Stream.bytesL16: ByteArray
    get() = read(read(Short.SIZE_BYTES).short.toUShort().toInt())
    set(bytes) {
        write(bytes, bytes.size.toShort().bytes)
    }

var Stream.bytesL32: ByteArray
    get() = read(read(Int.SIZE_BYTES).int)
    set(bytes) {
        write(bytes, bytes.size.bytes)
    }

var Stream.stringL8: String
    get() = bytesL8.decodeToString()
    set(string) {
        bytesL8 = string.encodeToByteArray()
    }

var Stream.stringL16: String
    get() = bytesL16.decodeToString()
    set(string) {
        bytesL16 = string.encodeToByteArray()
    }

