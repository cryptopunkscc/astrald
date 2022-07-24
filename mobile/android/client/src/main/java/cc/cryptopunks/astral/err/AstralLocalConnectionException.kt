package cc.cryptopunks.astral.err

import java.io.IOException

class AstralLocalConnectionException(cause: Throwable) :
    IOException("Cannot connect to local astral service.", cause)
