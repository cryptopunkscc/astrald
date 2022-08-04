# Astral android contacts

The library that provides access to astral contacts for android apps.

## Features

* Display contacts in android activity.
* Start contacts activity through uri.
* Select contact and return through activity results.

## API

> Start contacts activity

```kotlin
val intent = Intent(Intent.ACTION_VIEW, Uri.parse("astral://contacts"))
startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("astral://contacts")))
```

> Start contacts activity for selection results

```kotlin
// Create activity launcher
val contract = ActivityResultContracts.StartActivityForResult()
val launcher = registerForActivityResult(contract) { result ->
    // resolve contact id and name
    val identity = getStringExtra("identity")
    val name = getStringExtra("name")
}

// Prepare intent
val intent = Intent(Intent.ACTION_VIEW, Uri.parse("astral://contacts"))
intent.putBooleanExtra("select", true) // enable contact selection

// Start activity for results
launcher.launch(intent)
```
