# Warp Drive Android v1.0.0-draft

Few taps to share data with your mates anywhere you are.

# Features

The warpdrive features available on Android OS.

## `sender`

### `offer`

1. Open app that provides access to file and allows to share the file
    * via `android.intent.action.SEND` or `android.intent.action.SEND_MULTIPLE` intent.
2. Choose share option on file and select `warpdrive` app from list of available apps.
3. Choose and tap the recipient to send files offer.

* ShareFilesActivity displays list of available recipients.
    * Registers intent filter for `android.intent.action.SEND` and `android.intent.action.SEND_MULTIPLE`.
    * Allows sending file offer to selected recipient from list.

## `recipient`

### `offers`

1. Observe files offer notification in status bar.
2. Tap on the notification to see the details.

* FilesOfferNotification
    * display info about files offer.
    * can navigate to offer details.
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

* OfferDetailsActivity
    * Accept button
* DownloadNotification displays:
    * Files names.
    * Download progress

on sender site:

1. See the upload progress notification in status bar.
2. Wait for progress finish.
3. See downloaded files in warpdrive directory.

* UploadNotification displays:
    * Files names.
    * Upload progress.
* ContentResolverService:
    * astral service written kotlin exposes android feature not directly accessible in golang,
    * has access to android content resolver,
    * provides access to file bytes stream via uri.

# Components

List of components included in warpdrive android.

## Activities

### Send offer to contact

* ShareFilesActivity
    * register intent filter for `android.intent.action.SEND` and `android.intent.action.SEND_MULTIPLE`
    * get files uris from intent
    * fetch contacts from service
    * display list of fetched contacts
    * on contact click send offer request to service

### Display offer details

* OfferDetailsActivity
    * sender name
    * receive date time
    * list of files:
        * file name
        * file size
    * if offer is not accepted display:
        * accept button
            * on click call api accept

## Notifications

### New offer notification

* OfferNotification
    * display:
        * sender name
        * if single file:
            * file name
            * file size
        * if directory:
            * dir name
            * number of files
            * summary size
        * if group of files:
            * list of single files
    * on click navigate to OfferDetailsActivity

### Upload status notification

* ProgressNotification
    * one or group of files notifications for the same offer
    * display:
        * sender name
        * file name
        * download or upload progress
    * ongoing until progress not completed
    * on click navigate to OfferDetailsActivity

### API

```go
package warpdrive

type Notifier interface {
	Incoming(offer *Offer) error
	Outgoing(offer *Offer) error
	Start(offer *Offer, uri string, indeterminate bool) (err error)
	Progress(offer *Offer, uri string, progress int) (err error)
	Finish(offer *Offer, uri string, status string) (err error)
	FinishGroup(offer *Offer) error
}

type Offer struct{}
```

## Services

### Content resolver

* ContentResolverService
    * the astral service written in kotlin
    * have direct access android content resolver.
    * embedded and delivered with android node wrapper.
    * provides access to content via uri.

#### API

```go
package warpdrive

import "io"

type Resolver interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}

type Info struct{/*...*/}
```

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

```go
package notify

type Api interface {
	Create(channel Channel) error
	Notify(notifications ...Notification) error
}

type Notification struct {
	Id            int
	ChannelId     string
	ContentTitle  string
	SubText       string
	SmallIcon     string
	Ongoing       bool
	OnlyAlertOnce bool
	Group         string
	GroupSummary  bool
	*Progress
}

type Progress struct {
	Max           int
	Current       int
	Indeterminate bool
}

type Channel struct {
	Id         string
	Name       string
	Importance int
}
```

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
