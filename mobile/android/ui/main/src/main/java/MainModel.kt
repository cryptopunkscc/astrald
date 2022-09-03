package cc.cryptopunks.astral.ui.main

import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cc.cryptopunks.astral.node.astralIdentity
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch

class MainModel : ViewModel(), DefaultLifecycleObserver {

    val id = MutableStateFlow("")

    override fun onCreate(owner: LifecycleOwner) {
        viewModelScope.launch { loadId() }
    }

    private suspend fun loadId() {
        id.value = astralIdentity()
    }
}
