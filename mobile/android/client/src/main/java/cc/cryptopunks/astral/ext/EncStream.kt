package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncStream

fun EncStream.readMessage(): String? {
    val result = StringBuilder()
    val buffer = ByteArray(4096)
    var len: Int
    do {
        len = read(buffer)
        if (len > 0) result.append(String(buffer.copyOf(len)))
    } while (len == buffer.size)
    return when {
        len == -1 && result.isEmpty() -> null
        else -> result.toString()
    }
}

fun EncStream.readMessage(handle: (String) -> Unit): Boolean =
    readMessage()?.let(handle) != null


fun EncStream.readL64Bytes(): ByteArray =
    read(long.toInt())

fun EncStream.encodeL8(any: Any) {
    stringL8 = encoder.encode(any)
}

fun EncStream.encodeL16(any: Any) {
    stringL16 = encoder.encode(any)
}

fun <T> EncStream.decodeL8(type: Class<T>): T =
    encoder.decode(stringL8, type)

fun <T> EncStream.decodeL16(type: Class<T>): T =
    encoder.decode(stringL16, type)

inline fun <reified T> EncStream.decodeL8(): T =
    encoder.decode(stringL8, T::class.java)

inline fun <reified T> EncStream.decodeL16(): T =
    encoder.decode(stringL16, T::class.java)

inline fun <reified T> EncStream.decodeL16Array(): Array<T> =
    encoder.decodeArray(stringL16, T::class.java)
