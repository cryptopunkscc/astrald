# Warp Drive Android v1.0.0-draft

Few taps to share data with your mates anywhere you are.

# Use cases

> The sender is sending files offer to the recipient.

1. User opens app that provides access to file and allows to share the file
    * via `android.intent.action.SEND` or `android.intent.action.SEND_MULTIPLE` intent.
2. User chooses share option on file and selects `Warp Drive` app from list of available apps.
3. Application displays screen with available recipients.
4. User selects the recipient to send files offer.
5. Application sends file offer.
6. Application displays a snackbar with feedback about, success or failure.

#### ShareActivity

Receives share-files intent and allows sending file offer to recipient.

1. register intent filter for `android.intent.action.SEND` and `android.intent.action.SEND_MULTIPLE`;
2. get files uris from intent;
3. fetch contacts from service;
    * display error snackbar if failed.
    * pull down to refresh contacts.
4. display list of fetched contacts, for each item show:
    * Contact alias
    * 8 last characters of node id
5. on contact click send offer request to service;
    1. display progress bar.
    2. on finish display the snack bar with info:
        * Share delivered and waiting for approval.
        * Share accepted, the files are sending in background.
        * Sharing error - with error message.

---

> The recipient is notified about files offer.

1. Application receives files offer.
2. Application notifies the user about files offer.
3. User sees offer notification in status bar.

#### OfferNotification

Displays short details about files offer and navigates the user into offer details.

1. on click dismiss and navigate the user to OfferActivity.
3. can be swiped.
4. display info:
    * {Sender name} wants to share file(s) with you.
    * display file name if file is only one.
    * Name of root directory of files if is only one.
    * Number of files if more than one.
    * Summary size.

---

> The recipient navigates from notification to offer details.

1. User taps on the notification to see the details.
2. Application displays offer details screen.

#### OfferActivity

Loads and displays offer details.

1. get offer id from intent;
2. fetch offer from service;
3. displays offer details:
    * sender name
    * offer id
    * status
    * timestamp
    * list of files
        * file name
        * file size

---

> The recipient is accepting files offer from sender and downloading is started in background.

1. User sees the offer details screen.
2. User clicks the download button.
3. Application starts downloading files in background.
4. Application displays feedback about:
    * downloading started with success.
    * downloading failed with error.

#### OfferActivity

Displays download button. Starts downloading on click.

1. display download button if the files offer was received (not sent).
2. on button click start downloading in background
3. display snack bar with info
    * downloading started.
    * downloading failed - error message.

---

> The user is tracking file transferring status.

1. Application starts downloading files in background.
2. Application displays progress status, in:
    * silent notification updates in status bar.
    * live updates in offer details screen.

#### OfferNotification

Live updates about files transfer status.

* is ongoing and cannot be canceled.
* is silent, no sound, no buzz and no lights.
* on click navigates to OfferActivity.
* display info about files offer.
    * {Sender name} is sending file.
    * Name of current file.
    * Summary size.
    * Progress

#### OfferActivity

Live detailed updates about transfer status.

1. title - transfer in progress.
2. on activity resume, start listening the offer updates.
3. display offer update:
    * summary progress.
    * specific file progress.

---

> The user is notified about transfer finish.

1. Application finished file downloading, because:
    * is completed with success;
    * the error occurs;
2. Application notifies the user about finish, updating:
    * progress notification in status bar.
    * offer details screen, if is currently displayed.

#### OfferNotification

Displays information about transfer failure or success.

1. on click dismiss and navigate the user to OfferActivity.
2. can be swiped.
3. display info about transfer finish.
    * {Sender name} - transfer [succeed, failed]:
    * Name of the file if is only one.
    * Name of root directory of files if is only one.
    * Number of files if more than one.
    * Summary size.

#### OfferActivity

* OfferDetailsActivity
    * Accept button
* DownloadNotification displays:
    * Files names.
    * Download progress

---

## Services

Warp Drive UI for Android is a client that serves user interface and allows interacting with core background service
written in golang. This service requires access to some Android platform specific features, that can be exposed as
platform-specific astral services.

### Content resolver

Android provides standard access to file via internal content resolver api which is not directly accessible for golang
and needs to be exported via astral service, so the warpdrive core can access file later in background without help from
android client app.

* the astral service written in kotlin
* have direct access android content resolver.
* embedded and delivered with android node wrapper.
* provides access to content via uri.

#### API

Check the source file [api.go](/mobile/android/node/content/api.go)

#### Protocol

| </ `read` | </ `info` |
|-----------|-----------|
| <- uri    | <- uri    |
| -> bytes  | -> info   |
| <- ok     | <- ok     |

### Notifications

* NotificationService
    * the astral service written in kotlin.
    * have direct access android notification manager.
    * embedded and delivered with android node wrapper.
    * provides access to android notifications api.

#### API

Check the source file [api.go](/mobile/android/node/notify/api.go).

### Frames

| name          | short | format | representation |
|:--------------|:------|:-------|:---------------|
| Channel       | chan  | struct | Channel        | 
| Notifications | items | struct | []Notification | 

### Protocol

| </ `channel` | </ `notify` |
|--------------|-------------|
| <- chan      | <- items    |
| -> ok        | -> ok       |
