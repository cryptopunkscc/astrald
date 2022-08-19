package cc.cryptopunks.astral.client

import java.nio.ByteBuffer
import java.nio.ByteBuffer.allocate
import java.nio.ByteOrder.BIG_ENDIAN

private val byteOrder get() = BIG_ENDIAN

val Byte.bytes: ByteArray get() = ByteArray(1) { this }
val Short.bytes: ByteArray get() = allocate(2).order(byteOrder).putShort(this).array()
val Int.bytes: ByteArray get() = allocate(4).order(byteOrder).putInt(this).array()
val Long.bytes: ByteArray get() = allocate(8).order(byteOrder).putLong(this).array()

val ByteArray.byte: Byte get() = this[0]
val ByteArray.short: Short get() = ByteBuffer.wrap(this).order(byteOrder).short
val ByteArray.int: Int get() = ByteBuffer.wrap(this).order(byteOrder).int
val ByteArray.long: Long get() = ByteBuffer.wrap(this).order(byteOrder).long
