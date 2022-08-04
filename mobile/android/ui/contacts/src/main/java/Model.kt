package cc.cryptopunks.astral.ui.contacts

import android.app.Activity
import android.content.Intent
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cc.cryptopunks.astral.client.enc.encoder
import cc.cryptopunks.astral.client.enc.gson.GsonCoder
import cc.cryptopunks.astral.client.tcp.astralTcpNetwork
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch

class ContactsModel : ViewModel() {
    var selectable = true
    val loading = MutableStateFlow(false)
    val selected = MutableSharedFlow<Contact>(extraBufferCapacity = 1)
    val contacts = MutableStateFlow(listOf<Contact>())
}

fun ContactsModel.loadContacts() =
    viewModelScope.launch {
        try {
            loading.value = true
            contacts.value = astral.getContacts()
        } catch (e: Throwable) {
            e.printStackTrace()
        } finally {
            loading.value = false
        }
    }

fun ContactsModel.handleContactSelection(
    activity: Activity,
) = viewModelScope.launch {
    selected.collect { contact ->
        activity.setResult(Activity.RESULT_OK, Intent().apply {
            putExtra("identity", contact.id)
            putExtra("name", contact.name)
        })
        activity.finish()
    }
}

private val astral = astralTcpNetwork().encoder(GsonCoder())
