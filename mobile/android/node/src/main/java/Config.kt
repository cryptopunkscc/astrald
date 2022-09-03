package cc.cryptopunks.astral.node

import android.content.Context
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.io.File

data class Config(
    val alias: String = "",
)

val EmptyConfig = Config()

val astralConfig: StateFlow<Config> get() = state

private val state = MutableStateFlow(EmptyConfig)

private val File.astralConfig get() = resolve("astrald.conf")

private val Context.astralConfig get() = astralDir.nodeDir.astralConfig

fun Context.loadAstralConfig() {
    if (state.value != EmptyConfig) return
    val file = astralConfig
    if (file.exists()) try {
        file.readLines()
            .find { it.startsWith("alias") }
            ?.split(":")?.get(1)
            ?.trim()?.let { alias ->
                state.value = Config(
                    alias = alias
                )
            }
    } catch (e: Throwable) {
        println("Cannot load config")
        e.printStackTrace()
    }
}

fun Context.setAstralConfig(config: Config) {
    try {
        astralConfig.writeText("alias: ${config.alias}")
        state.value = config
    } catch (e: Throwable) {
        println("Cannot set config")
        e.printStackTrace()
    }
}
