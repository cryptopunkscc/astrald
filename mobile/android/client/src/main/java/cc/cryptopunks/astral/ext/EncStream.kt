package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncStream

fun EncStream.encodeL8(any: Any) {
    stringL8 = encoder.encode(any)
}

fun EncStream.encodeL16(any: Any) {
    stringL16 = encoder.encode(any)
}

fun <T> EncStream.decodeMessage(type: Class<T>): T =
    encoder.decode(readMessage().orEmpty(), type)

fun <T> EncStream.decodeL8(type: Class<T>): T =
    encoder.decode(stringL8, type)

fun <T> EncStream.decodeL16(type: Class<T>): T =
    encoder.decode(stringL16, type)

inline fun <reified T> EncStream.decodeMessage(): T =
    encoder.decode(readMessage().orEmpty(), T::class.java)

inline fun <reified T> EncStream.decodeArray(): Array<T> =
    encoder.decodeArray(readMessage().orEmpty(), T::class.java)

inline fun <reified K, reified V> EncStream.decodeMap(): Map<K, V> =
    encoder.decodeMap(readMessage().orEmpty(), K::class.java, V::class.java)

inline fun <reified T> EncStream.decodeL8(): T =
    encoder.decode(stringL8, T::class.java)

inline fun <reified T> EncStream.decodeL16(): T =
    encoder.decode(stringL16, T::class.java)

inline fun <reified T> EncStream.decodeL16Array(): Array<T> =
    encoder.decodeArray(stringL16, T::class.java)
