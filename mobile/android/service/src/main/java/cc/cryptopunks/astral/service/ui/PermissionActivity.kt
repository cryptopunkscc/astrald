package cc.cryptopunks.astral.service.ui

import android.Manifest.permission.WRITE_EXTERNAL_STORAGE
import android.os.Bundle
import android.widget.Button
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import cc.cryptopunks.astral.service.R

class PermissionActivity : AppCompatActivity() {

    private val requestPermission = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted: Boolean ->
        if (isGranted) finish()
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.permission_view)
        findViewById<Button>(R.id.grantPermission).setOnClickListener {
            requestPermission.launch(WRITE_EXTERNAL_STORAGE)
        }
    }
}
