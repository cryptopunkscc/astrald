package cc.cryptopunks.astral.ui.main

import android.content.Intent
import android.widget.Toast
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.material.Button
import androidx.compose.material.DropdownMenu
import androidx.compose.material.DropdownMenuItem
import androidx.compose.material.Icon
import androidx.compose.material.IconButton
import androidx.compose.material.Scaffold
import androidx.compose.material.Text
import androidx.compose.material.TopAppBar
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import cc.cryptopunks.astral.intent.startAstralService
import cc.cryptopunks.astral.intent.stopAstralService
import cc.cryptopunks.astral.ui.log.LogModel
import cc.cryptopunks.astral.ui.log.LogView
import cc.cryptopunks.astral.node.AstralStatus
import cc.cryptopunks.astral.node.astralStatus
import kotlinx.coroutines.launch

@Composable
internal fun MainView(
    mainModel: MainModel,
    logModel: LogModel,
    menuItems: Map<String, Intent>,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("Astral")
                },
                actions = {
                    if (menuItems.isNotEmpty()) {
                        var displayMenu by remember { mutableStateOf(false) }
                        IconButton(
                            onClick = { displayMenu = !displayMenu }
                        ) {
                            Icon(Icons.Default.MoreVert, "")
                        }
                        DropdownMenu(
                            expanded = displayMenu,
                            onDismissRequest = { displayMenu = false }
                        ) {
                            val context = LocalContext.current
                            menuItems.forEach { (name, intent) ->
                                println("menu item: $name")
                                DropdownMenuItem(
                                    onClick = {
                                        context.startActivity(intent)
                                        displayMenu = false
                                    }
                                ) {
                                    Text(name)
                                }
                            }
                        }
                    }
                }
            )
        }
    ) {
        Column {
            val context = LocalContext.current
            val scope = rememberCoroutineScope()

            // astral identity
            val id by mainModel.id.collectAsState()
            Text(
                text = id,
                modifier = Modifier.pointerInput(Unit) {
                    detectTapGestures(onLongPress = {
                        context.copyToClipboard("identity", id)
                        Toast.makeText(context,
                            "Id copied to clipboard.",
                            Toast.LENGTH_SHORT).show()
                    })
                }
            )

            // logs
            LogView(
                modifier = Modifier.weight(1f),
                logModel = logModel,
            )

            // buttons
            Row {
                val status by astralStatus.collectAsState()
                Button(
                    enabled = status == AstralStatus.Stopped,
                    onClick = { context.startAstralService() },
                ) {
                    Text("start node")
                }
                Button(
                    enabled = status == AstralStatus.Started,
                    onClick = { context.stopAstralService() },
                ) {
                    Text("stop node")
                }
                var fixLogButtonEnabled by remember { mutableStateOf(true) }
                Button(
                    enabled = fixLogButtonEnabled,
                    onClick = {
                        scope.launch {
                            fixLogButtonEnabled = false
                            logModel.clearLogs()
                            fixLogButtonEnabled = true
                        }
                    }
                ) {
                    Text("clear logs")
                }
            }
        }
    }
}
