package cc.cryptopunks.astral.service.content

import android.content.Context
import cc.cryptopunks.astral.ext.byte
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.ext.stringL8
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.io.outputStream
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch

private object Port {
    const val READ = "sys/content"
    const val INFO = "sys/content/info"
}

private val astral = astralTcpNetwork(GsonCoder())

suspend fun Context.startContentResolverService() {
    coroutineScope {
        Adapter(contentResolver).run {
            launch { handleRead() }
            launch { handleInfo() }
        }
    }
}

private suspend fun Content.Resolver.handleRead() {
    astral.register(Port.READ) {
        val uri = stringL8
        byte = 0
        val output = outputStream()
        reader(uri).copyTo(output)
        output.flush()
    }
}

private suspend fun Content.Resolver.handleInfo() {
    astral.register(Port.INFO) {
        val uri = stringL8
        val info = info(uri)
        val data = encoder.encode(arrayOf(info))
        write(data.toByteArray())
    }
}
