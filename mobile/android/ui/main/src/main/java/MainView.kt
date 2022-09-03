package cc.cryptopunks.astral.ui.main

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import androidx.lifecycle.viewmodel.compose.viewModel
import cc.cryptopunks.astral.intent.startAstralService

@Composable
internal fun MainView(
    mainModel: MainModel = viewModel(),
) {
    val context = LocalContext.current
    val initialized by mainModel.initialized.collectAsState()
    if (!initialized) ConfigView()
    else {
        context.startAstralService()
        DashboardView(
            menuItems = context.actionIntents()
        )
    }
}
