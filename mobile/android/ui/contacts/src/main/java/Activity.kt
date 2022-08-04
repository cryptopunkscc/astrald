package cc.cryptopunks.astral.ui.contacts

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import cc.cryptopunks.astral.theme.AstralTheme

class ContactsActivity : ComponentActivity() {

    private val model by viewModels<ContactsModel>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        model.selectable = intent.getBooleanExtra("select", false)
        model.handleContactSelection(this)
        setContent {
            AstralTheme {
                MainView(model)
            }
        }
    }
}
