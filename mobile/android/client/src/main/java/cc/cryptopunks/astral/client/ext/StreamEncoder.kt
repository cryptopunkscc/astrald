package cc.cryptopunks.astral.client.ext

import cc.cryptopunks.astral.client.enc.StreamEncoder

fun StreamEncoder.encodeL8(any: Any) {
    string8 = encode(any)
}

fun StreamEncoder.encodeL16(any: Any) {
    string16 = encode(any)
}

fun StreamEncoder.encodeL32(any: Any) {
    string32 = encode(any)
}

fun StreamEncoder.encodeLine(any: Any) {
    write((encode(any) + "\n").encodeToByteArray())
}

fun <T> StreamEncoder.decode8(type: Class<T>): T =
    decode(string8, type)

fun <T> StreamEncoder.decode16(type: Class<T>): T =
    decode(string16, type)

fun <T> StreamEncoder.decode32(type: Class<T>): T =
    decode(string32, type)

fun <T> StreamEncoder.decodeMessage(type: Class<T>): T =
    decode(readMessage().orEmpty(), type)

inline fun <reified T> StreamEncoder.decodeMessage(): T =
    decode(readMessage().orEmpty(), T::class.java)

inline fun <reified T> StreamEncoder.decodeList(): List<T> =
    decodeList(readMessage().orEmpty(), T::class.java)

inline fun <reified K, reified V> StreamEncoder.decodeMap(): Map<K, V> =
    decodeMap(readMessage().orEmpty(), K::class.java, V::class.java)

inline fun <reified T> StreamEncoder.decode8(): T =
    decode(string8, T::class.java)

inline fun <reified T> StreamEncoder.decode16(): T =
    decode(string16, T::class.java)

inline fun <reified T> StreamEncoder.decode32(): T =
    decode(string32, T::class.java)

inline fun <reified T> StreamEncoder.decodeList8(): List<T> =
    decodeList(string8, T::class.java)

inline fun <reified T> StreamEncoder.decodeList16(): List<T> =
    decodeList(string16, T::class.java)

inline fun <reified T> StreamEncoder.decodeList32(): List<T> =
    decodeList(string32, T::class.java)
