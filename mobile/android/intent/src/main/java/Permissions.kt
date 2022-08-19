package cc.cryptopunks.astral.intent

import android.content.Context
import android.content.pm.PackageInfo
import android.content.pm.PackageManager

fun Context.hasPermissions(vararg permissions: String): Boolean =
    filterMissingPermissions(*permissions).isEmpty()

fun Context.filterMissingPermissions(vararg permissions: String): Array<String> {
    val info: PackageInfo? = packageManager
        .getInstalledPackages(PackageManager.GET_PERMISSIONS)
        .find { info -> info.packageName == astralPackageName }

    requireNotNull(info)

    return permissions.filter { permission ->
        val index = info.requestedPermissions.indexOf(permission)
        if (index < 0) true
        else info.requestedPermissionsFlags[index] and PackageInfo.REQUESTED_PERMISSION_GRANTED == 0
    }.toTypedArray()
}
