package cc.cryptopunks.astral.theme

import androidx.compose.material.MaterialTheme
import androidx.compose.material.lightColors
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

@Composable
fun AstralTheme(
    content: @Composable () -> Unit,
) = MaterialTheme(
    colors = lightColors(
        primary = Color(0xDE000000),
        primaryVariant = Color(0xFF000000),
    ),
    content = content
)
