package cc.cryptopunks.astral.service.content

import android.content.Context
import cc.cryptopunks.astral.ext.byte
import cc.cryptopunks.astral.ext.register
import cc.cryptopunks.astral.ext.stringL8
import cc.cryptopunks.astral.gson.GsonCoder
import cc.cryptopunks.astral.tcp.astralTcpNetwork
import kotlinx.coroutines.launch
import kotlinx.coroutines.supervisorScope

private object Port {
    const val READ = "sys/content"
    const val INFO = "sys/content/info"
}

private val astral = astralTcpNetwork(GsonCoder())

suspend fun Context.startContentResolverService() {
    supervisorScope {
        Adapter(contentResolver).run {
            launch {
                handleRead()
                println("Finishing read")
            }
            launch {
                handleInfo()
                println("Finishing info")
            }
        }
    }
}

private suspend fun Content.Resolver.handleRead() = supervisorScope {
    try {
        astral.register(Port.READ) {
            println("New read files request")
            val uri = stringL8
            byte = 0
            println()
            println("Start coping files")
            try {
                val file = reader(uri)
                file.copyTo(output)
                println("Flushing files")
                output.flush()
                println("Flushed successfully")
            } catch (e: Throwable) {
                println("Cannot copy file ${e.localizedMessage}")
                e.printStackTrace()
            }
        }
    } catch (e: Throwable) {
        println("Finish handling read")
        e.printStackTrace()
    }
}

private suspend fun Content.Resolver.handleInfo() {
    try {
        astral.register(Port.INFO) {
            val uri = stringL8
            try {
                val info = info(uri)
                val data = encoder.encode(arrayOf(info))
                write(data.toByteArray())
            } catch (e: Throwable) {
                println("Cannot send info ${e.localizedMessage}")
                e.printStackTrace()
            }
        }
    } catch (e: Throwable) {
        println("Finish handling info")
        e.printStackTrace()
    }
}
