package cc.cryptopunks.astral.ui.log

import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cc.cryptopunks.astral.node.startTime
import kotlinx.coroutines.Job
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.scan
import kotlinx.coroutines.launch
import java.util.*

class LogModel : ViewModel(), DefaultLifecycleObserver {

    private var job: Job? = null

    val log = MutableStateFlow(emptyList<String>())

    override fun onCreate(owner: LifecycleOwner) {
        startLogs()
    }

    suspend fun clearLogs() {
        stopLogs()
        clearLogcatProcess()
        startLogs()
    }

    private fun startLogs() {
        job == null || return
        job = viewModelScope.launch {
            logcatFlow(Date(startTime)).formatAstralLogs().throttle(Config.LogThrottle)
                .scan(emptyList<String>()) { acc, next ->
                    (acc + next).run {
                        if (size <= Config.MaxLogLines) this
                        else drop(size - Config.MaxLogLines)
                    }
                }
                .collect { lines -> log.value = lines }
        }
    }

    private suspend fun stopLogs() {
        job?.cancelAndJoin()
        job = null
    }
}
