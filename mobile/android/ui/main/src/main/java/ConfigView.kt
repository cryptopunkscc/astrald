package cc.cryptopunks.astral.ui.main

import android.os.Build
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.material.Button
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Scaffold
import androidx.compose.material.Text
import androidx.compose.material.TextField
import androidx.compose.material.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import cc.cryptopunks.astral.theme.AstralTheme

@Preview
@Composable
internal fun ConfigPreview() = AstralTheme {
    ConfigView()
}

@Composable
internal fun ConfigView(
    model: MainModel = viewModel(),
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("Configure Astral")
                },
            )
        }
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center,
        ) {
            Text(
                text = "Setup astral node alias",
                textAlign = TextAlign.Center,
                style = MaterialTheme.typography.h5,
            )
            Spacer(Modifier.height(32.dp))
            var alias by remember { mutableStateOf(defaultDeviceName()) }
            TextField(
                value = alias,
                onValueChange = { alias = it },
                placeholder = { Text("node alias")}
            )
            val context = LocalContext.current
            Spacer(Modifier.height(32.dp))
            Button(
                onClick = {
                    context.astralConfig.writeText("alias: $alias")
                    model.initialized.value = true
                }
            ) {
                Text("setup")
            }
        }
    }
}

private fun defaultDeviceName(): String {
    val manufacturer: String = Build.MANUFACTURER
    val model: String = Build.MODEL
    return if (model.startsWith(manufacturer))
        model.replaceFirstChar { it.uppercase() } else
        manufacturer.replaceFirstChar { it.uppercase() } + " " + model
}
