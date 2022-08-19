package cc.cryptopunks.astral.intent

import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.os.Build

// service

val astralServiceIntent
    get() = Intent().apply {
        component = ComponentName(astralPackageName, astralServiceClassName)
    }

fun Context.startAstralService(): ComponentName? = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O)
    startForegroundService(astralServiceIntent) else
    startService(astralServiceIntent)

fun Context.stopAstralService() = stopService(astralServiceIntent)

// main activity

val astralActivityIntent get() = astralMainActivityUri.intent()

fun Context.startAstralActivity() = startActivity(astralActivityIntent)

// permissions activity

fun permissionsActivityIntent(text: String, vararg permissions: String) =
    astralPermissionsActivityUri.intent {
        putExtra("text", text)
        putExtra("request", permissions)
    }
