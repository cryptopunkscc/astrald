# Android warpdrive client

# Features

## `sender`

### `offer`

1. Open app that provides access to file and allows to share the file
    * via `android.intent.action.SEND` or `android.intent.action.SEND_MULTIPLE` intent.
2. Choose share option on file and select `warpdrive` app from list of available apps.
3. Choose and tap the recipient to send files offer.

#### Activity

* ShareFilesActivity displays list of available recipients.
    * Registers intent filter for `android.intent.action.SEND` and `android.intent.action.SEND_MULTIPLE`.
    * Allows sending file offer to selected recipient from list.

## `recipient`

### `offers`

1. Observe files offer notification in status bar.
2. Tap on the notification to see the details.

#### Notification

* FilesOfferNotification
    * display info about files offer.
    * can navigate to offer details.

#### Activity

* OfferDetailsActivity
    * displays offer details like:
        * Sender name
        * List of files
            * file name
            * file size
        * Timestamp

### `accept`

1. Open the offer details screen.
2. Click accept button.
3. See the download progress notification in status bar.

on sender site:

1. See the upload progress notification in status bar.

#### Notification

* UploadNotification displays:
    * Files names.
    * Upload progress.
* DownloadNotification displays:
    * Files names.
    * Download progress

#### Activity

* OfferDetailsActivity
    * Accept button

#### Service

* ProgressNotificationService
    * Exported
    * Require astral port as argument
    * Manages downloading or uploading progress
    * Connects to warpdrive service for progress for receiving progress updates
    * Displays progress notification.