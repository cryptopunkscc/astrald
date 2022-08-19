package cc.cryptopunks.astral.client.enc.gson

import cc.cryptopunks.astral.client.enc.Encoder
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

val defaultGson by lazy {
    Gson().newBuilder()
        .setFieldNamingStrategy { it.name.replaceFirstChar(Char::uppercase) }
        .create()
}

val coder by lazy {
    GsonCoder(defaultGson)
}

class GsonCoder(
    val gson: Gson = defaultGson,
) : Encoder {
    override fun encode(any: Any): String = gson.toJson(any)
    override fun <T> decode(string: String, type: Class<T>): T = gson.fromJson(string, type)

    override fun <T> decodeList(string: String, type: Class<T>): List<T> = when {
        string.isBlank() -> emptyList()
        else -> gson.fromJson<Array<T>>(string, TypeToken.getArray(
            TypeToken.get(type).type).type
        )?.toList() ?: emptyList()
    }

    override fun <K, V> decodeMap(string: String, key: Class<K>, value: Class<V>): Map<K, V> =
        gson.fromJson(string, TypeToken.getParameterized(
            TypeToken.get(Map::class.java).type, key, value
        ).type)
}
