package cc.cryptopunks.astral.android

typealias Call = (String) -> Unit
typealias Get = (String) -> ByteArray
typealias Read = (String, Write) -> Unit
typealias Write = (ByteArray) -> Long

abstract class AndroidApi {
    protected abstract val method: Methods
    fun call(path: String, arg: String) = method.call.getValue(path)(arg)
    fun get(path: String, arg: String): ByteArray = method.get.getValue(path)(arg)
    fun read(path: String, arg: String, write: Write) = method.read.getValue(path)(arg, write)

    data class Methods(
        val call: Map<String, Call> = emptyMap(),
        val get: Map<String, Get> = emptyMap(),
        val read: Map<String, Read> = emptyMap(),
    )
}

operator fun AndroidApi.Methods.plus(
    other: AndroidApi.Methods,
) = AndroidApi.Methods(
    call = call + other.call,
    get = get + other.get,
    read = read + other.read,
)
