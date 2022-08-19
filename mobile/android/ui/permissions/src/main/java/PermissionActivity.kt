package cc.cryptopunks.astral.ui.permissions

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.Button
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Scaffold
import androidx.compose.material.Text
import androidx.compose.material.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import cc.cryptopunks.astral.theme.AstralTheme

class PermissionActivity : ComponentActivity() {

    private companion object {
        const val requestKeyPermissions = "request"
        const val requestKeyText = "text"
        const val resultKeyRejected = "rejected"
    }

    private var required = emptyArray<String>()

    private val request = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { result ->
        val rejected = result.filterValues { !it }.keys.toTypedArray()
        val intent = Intent().putExtra(resultKeyRejected, rejected)
        setResult(RESULT_OK, intent)
        finish()
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // resolve requested permissions
        required = intent.getStringArrayExtra(requestKeyPermissions) ?: return
        required.isNotEmpty() || return

        val message = intent.getStringExtra(requestKeyText) ?: ""
        setContent {
            AstralTheme {
                PermissionsView(message) {
                    request.launch(required)
                }
            }
        }
    }
}

@Preview
@Composable
private fun PermissionsPreview() {
    AstralTheme {
        PermissionsView(message = "Test message") {

        }
    }
}

@Composable
private fun PermissionsView(
    message: String,
    onClick: () -> Unit,
) = Scaffold(
    topBar = {
        TopAppBar(
            title = {
                Text("Permission required")
            }
        )
    }
) {
    Column(
        Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        // setup permissions rationale
        Text(
            text = message,
            modifier = Modifier.padding(64.dp),
            textAlign = TextAlign.Center,
            style = MaterialTheme.typography.h5,
        )
        // setup grant permissions button
        Button(
            content = {
                Text("grant permissions")
            },
            onClick = onClick,
        )
    }
}
