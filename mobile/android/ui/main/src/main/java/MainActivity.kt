package cc.cryptopunks.astral.ui.main

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import cc.cryptopunks.astral.node.loadAstralConfig
import cc.cryptopunks.astral.theme.AstralTheme
import cc.cryptopunks.astral.ui.log.LogModel

class MainActivity : ComponentActivity() {

    private val mainModel by viewModels<MainModel>()
    private val logModel by viewModels<LogModel>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        loadAstralConfig()
        lifecycle.apply {
            addObserver(mainModel)
            addObserver(logModel)
        }
        setContent {
            AstralTheme {
                MainView()
            }
        }
    }
}
