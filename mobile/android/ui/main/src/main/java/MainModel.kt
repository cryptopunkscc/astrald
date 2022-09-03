package cc.cryptopunks.astral.ui.main

import android.app.Application
import android.content.Context
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.viewModelScope
import cc.cryptopunks.astral.node.astralConfig
import cc.cryptopunks.astral.node.astralDir
import cc.cryptopunks.astral.node.astralIdentity
import cc.cryptopunks.astral.node.nodeDir
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch

class MainModel(
    private val context: Application,
) : AndroidViewModel(context),
    DefaultLifecycleObserver {

    val initialized = MutableStateFlow(false)
    val id = MutableStateFlow("")

    override fun onCreate(owner: LifecycleOwner) {
        initialized.value = context.astralConfig.exists()
        viewModelScope.launch { loadId() }
    }

    private suspend fun loadId() {
        id.value = astralIdentity()
    }
}

val Context.astralConfig get() = astralDir.nodeDir.astralConfig
