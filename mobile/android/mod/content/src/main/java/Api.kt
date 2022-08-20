package cc.cryptopunks.astral.mod.content

internal object Content {
    data class Info(
        val uri: String,
        val size: Long,
        val name: String = "",
        val mime: String = "",
    )

    const val read = "android/content"
    const val info = "android/content/info"
}
