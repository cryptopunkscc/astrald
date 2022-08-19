package cc.cryptopunks.astral.node

import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothSocket
import astral.Writer
import java.io.InputStream

internal class Bluetooth : astral.Bluetooth {

    private val name = "AstralNode"

    override fun dial(address: ByteArray): Socket {
        address.reverse()
        println("Dial to ${address.toHex()}")

        val adapter: BluetoothAdapter = BluetoothAdapter.getDefaultAdapter()

        val remoteDevice = adapter.getRemoteDevice(address)

        val socket = remoteDevice.createInsecureRfcommSocket(1)

        adapter.cancelDiscovery()

        return try {
            socket.connect()
            println("Connected to ${address.toHex()}")

            Socket(
                outbound = true,
                btSocket = socket,
                localAddress = null,
            )
        } catch (e: Throwable) {
            socket.close()
            e.printStackTrace()
            throw e
        }
    }

    class Socket(
        private val localAddress: String?,
        private val outbound: Boolean,
        private val btSocket: BluetoothSocket,
    ) : astral.BluetoothSocket {


        override fun localAddr(): String? {
            println("Getting local address: $localAddress")
            return localAddress
        }

        override fun outbound(): Boolean {
            println("Getting outbound: $outbound")
            return outbound
        }

        override fun remoteAddr(): String? {
            return btSocket.remoteDevice.address.also {
                println("Getting remote address $it")
            }
        }

        override fun close() {
            println("Closing connection")
            btSocket.outputStream.flush()
            btSocket.close()
            println("Connection closed")
        }

        override fun read(writer: Writer) {
            btSocket.inputStream.copyTo(writer)
        }

        override fun write(bytes: ByteArray): Long {
            btSocket.outputStream.write(bytes)
            return bytes.size.toLong()
        }
    }
}

private fun InputStream.copyTo(writer: Writer) {
    val size = 16 * 1024
    val buffer = ByteArray(size)
    var len: Int
    while (true) {
        len = read(buffer)
        when (len) {
            size -> writer.write(buffer)
            -1 -> break
            else -> writer.write(buffer.copyOf(len))
        }
    }
}

private fun BluetoothDevice.createInsecureRfcommSocket(port: Int): BluetoothSocket {
    val m = javaClass.getMethod("createInsecureRfcommSocket", Int::class.javaPrimitiveType)
    val soc = m.invoke(this, port)
    return soc as BluetoothSocket
}

private fun ByteArray.toHex(): String =
    joinToString(separator = ":") { eachByte -> "%02x".format(eachByte) }
