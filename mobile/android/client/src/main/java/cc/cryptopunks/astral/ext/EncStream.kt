package cc.cryptopunks.astral.ext

import cc.cryptopunks.astral.enc.EncStream

fun EncStream.encodeL8(any: Any) {
    stringL8 = encode(any)
}

fun EncStream.encodeL16(any: Any) {
    stringL16 = encode(any)
}

fun <T> EncStream.decodeMessage(type: Class<T>): T =
    decode(readMessage().orEmpty(), type)

fun <T> EncStream.decodeL8(type: Class<T>): T =
    decode(stringL8, type)

fun <T> EncStream.decodeL16(type: Class<T>): T =
    decode(stringL16, type)

inline fun <reified T> EncStream.decodeMessage(): T =
    decode(readMessage().orEmpty(), T::class.java)

inline fun <reified T> EncStream.decodeList(): List<T> =
    decodeList(readMessage().orEmpty(), T::class.java)


inline fun <reified K, reified V> EncStream.decodeMap(): Map<K, V> =
    decodeMap(readMessage().orEmpty(), K::class.java, V::class.java)

inline fun <reified T> EncStream.decodeL8(): T =
    decode(stringL8, T::class.java)

inline fun <reified T> EncStream.decodeL16(): T =
    decode(stringL16, T::class.java)

inline fun <reified T> EncStream.decodeL16List(): List<T> =
    decodeList(stringL16, T::class.java)
