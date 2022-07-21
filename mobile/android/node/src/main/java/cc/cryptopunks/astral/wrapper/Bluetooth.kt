package cc.cryptopunks.astral.wrapper

import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothServerSocket
import android.bluetooth.BluetoothSocket
import astral.Writer
import java.io.InputStream
import java.util.*

internal class Bluetooth : astral.Bluetooth {

    private val name = "AstralNode"

    private val uuid: UUID = UUID.fromString("00001101-0000-1000-8000-00805f9b34fb")
//    private val uuid: UUID = UUID.fromString("8ce255c0-200a-11e0-ac64-0800200c9a66")
    private val adapter: BluetoothAdapter = BluetoothAdapter.getDefaultAdapter()
    private val localAddress = adapter.address

    override fun address(): String = localAddress

    override fun listen(): ServerSocket {
        println("Creating server socket with address $localAddress")
//        val socket = adapter.listenUsingInsecureRfcommOn(1)
        val socket = adapter.listenUsingInsecureRfcommWithServiceRecord(name, uuid)
        
        return ServerSocket(
            localAddress = localAddress,
            btServerSocket = socket,
        )
    }

    override fun dial(address: ByteArray): Socket {
        address.reverse()
        println("Dial to ${address.toHex()}")

        val remoteDevice = adapter.getRemoteDevice(address)

        val socket = remoteDevice.createInsecureRfcommSocket(1)

        adapter.cancelDiscovery()

        return try {
            socket.connect()
            println("Connected to ${address.toHex()}")

            Socket(
                localAddress = localAddress,
                outbound = true,
                btSocket = socket

            )
        } catch (e: Throwable) {
            socket.close()
            e.printStackTrace()
            throw e
        }
    }


    class ServerSocket(
        val localAddress: String,
        private val btServerSocket: BluetoothServerSocket
    ) : astral.BluetoothPort {

        override fun accept(): astral.BluetoothSocket {
            println("BT accepting connection")
            try {
                val socket = btServerSocket.accept()
                return Socket(
                    localAddress = localAddress,
                    outbound = false,
                    btSocket = socket
                ).apply {
                    println("BT accepting connection from ${remoteAddr()}")
                }
            } catch (e: Throwable) {
                e.printStackTrace()
                throw e
            }
        }

        override fun close() {
            println("BT closing server socket with address $localAddress")
            btServerSocket.close()
        }
    }

    class Socket(
        private val localAddress: String,
        private val outbound: Boolean,
        private val btSocket: BluetoothSocket
    ) : astral.BluetoothSocket {



        override fun localAddr(): String {
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
    while(true) {
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


private fun BluetoothAdapter.listenUsingInsecureRfcommOn(port: Int): BluetoothServerSocket {
    val m = javaClass.getMethod("listenUsingInsecureRfcommOn", Int::class.javaPrimitiveType)
    val soc = m.invoke(this, port)
    return soc as BluetoothServerSocket
}

private fun ByteArray.toHex(): String =
    joinToString(separator = ":") { eachByte -> "%02x".format(eachByte) }
