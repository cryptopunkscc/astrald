package cc.cryptopunks.astral.ui.main

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import cc.cryptopunks.astral.intent.startAstralService
import cc.cryptopunks.astral.node.EmptyConfig
import cc.cryptopunks.astral.node.astralConfig

@Composable
internal fun MainView() {
    val config by astralConfig.collectAsState()
    if (config == EmptyConfig) ConfigView()
    else {
        val context = LocalContext.current
        context.startAstralService()
        DashboardView(
            menuItems = context.actionIntents()
        )
    }
}
