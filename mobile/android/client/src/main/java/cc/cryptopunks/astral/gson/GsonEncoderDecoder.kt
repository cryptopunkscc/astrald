package cc.cryptopunks.astral.gson

import cc.cryptopunks.astral.enc.Encoder
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

val defaultGson by lazy {
    Gson().newBuilder()
        .setFieldNamingStrategy { it.name.replaceFirstChar(Char::uppercase) }
        .create()
}

class GsonCoder(
    private val gson: Gson = defaultGson,
) : Encoder {
    override fun encode(any: Any): String = gson.toJson(any)
    override fun <T> decode(string: String, type: Class<T>): T = gson.fromJson(string, type)
    override fun <T> decodeArray(string: String, type: Class<T>): Array<T> =
        gson.fromJson(string, TypeToken.getArray(TypeToken.get(type).type).type)
}
