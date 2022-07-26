package cc.cryptopunks.astral.service.ui

import android.os.Build
import android.os.Bundle
import android.util.Log
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import cc.cryptopunks.astral.service.AstralService
import cc.cryptopunks.astral.service.Config
import cc.cryptopunks.astral.service.R
import cc.cryptopunks.astral.service.util.throttle
import cc.cryptopunks.astral.wrapper.AstralStatus
import cc.cryptopunks.astral.wrapper.astralIdentity
import cc.cryptopunks.astral.wrapper.astralStatus
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.MainScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : AppCompatActivity(), CoroutineScope by MainScope() {

    private val nodeIdTextView: TextView by lazy { findViewById(R.id.nodeIdTestView) }
    private val logView: RecyclerView by lazy { findViewById(R.id.list) }
    private val startButton: TextView by lazy { findViewById(R.id.startServiceButton) }
    private val killButton: TextView by lazy { findViewById(R.id.killServiceButton) }
    private val fixLogButton: TextView by lazy { findViewById(R.id.fixLogButton) }
    private val serviceIntent by lazy { AstralService.intent(this) }
    private val linearLayoutManager by lazy { LinearLayoutManager(this) }
    private val logAdapter = LogAdapter()
    private var logcatJob: Job? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        startAstralService()
        logView.apply {
            layoutManager = linearLayoutManager
            adapter = logAdapter
        }
        launch { nodeIdTextView.text = astralIdentity() }
        launch {
            astralStatus.collect { status ->
                startButton.isEnabled = status == AstralStatus.Stopped
                killButton.isEnabled = status == AstralStatus.Started
            }
        }
        nodeIdTextView.setOnLongClickListener { copyNodeId(); true }
        killButton.setOnClickListener { stopAstralService() }
        startButton.setOnClickListener { startAstralService() }
        fixLogButton.setOnClickListener { clearLogcatProcess() }

    }

    override fun onStart() {
        super.onStart()
        logcatJob = launch {
            logcatCacheFlow().throttle(Config.LogThrottle).collect { log ->
                withContext(Dispatchers.Main) {
                    logAdapter.run {
                        val startPos = lines.size
                        val lastVisiblePos = linearLayoutManager.findLastVisibleItemPosition()
                        lines += log
                        notifyItemRangeChanged(startPos, log.size)
                        if (lastVisiblePos > startPos - 2) logView.post {
                            logView.scrollToPosition(lines.size - 1)
                        }
                    }
                }
            }
        }
    }

    override fun onStop() {
        super.onStop()
        logcatJob?.cancel()
    }

    override fun onDestroy() {
        Log.d("MainActivity", "onDestroy")
        cancel()
        super.onDestroy()
    }

    private fun copyNodeId() {
        copyToClipboard(nodeIdTextView.text.toString())
        Toast.makeText(this, "Id copied to clipboard.", Toast.LENGTH_SHORT).show()
    }

    private fun startAstralService() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O)
            startForegroundService(serviceIntent) else
            startService(serviceIntent)
    }

    private fun stopAstralService() {
        stopService(serviceIntent)
    }
}
