# Virtual FS UI

Is a warpdrive client that serves user interface through virtual file system. Perhaps any file manager can be used to
communicate with warpdrive service through virtual file storage interface. Additionally, the user have to notified about
requests from peers, this feature requires platform specific adapter for notification management, and tray indicator.

## Features

## `sender`

### `send`

1. Open directory with mounted warpdrive peers;
2. Localize proper directory related to receiver;
3. Copy-paste a file into the directory;
4. The feedback about progress or failure will be displayed as notification.

### `incoming`

1. The warp service client have to be run and connected to warpdrive service;
2. Client receives incoming file request through the service;
3. Client displays the system notification about incoming file;
4. The notification can navigate to the directory with list of incoming file requests.
    * Each request is represented as a single file.

## `recipient`

### `accept`

1. Open directory that contains incoming file requests.
2. Move request file into `accepted` directory

### `reject`

1. Open directory that contains incoming file requests.
2. Move request file into `rejected` directory

### `update`

1. Open `peers` file.
2. Replace value near required peer into:
    1. clear value - default
    2. `trust`
    3. `block`

## Directory structure

* `warpdrive/`
    * `incoming/` - Incoming file requests management.
        * `accepted/` - Contains accepted files requests.
        * `rejected/` - Contains rejected files requests.
        * `peers` - list of trusted, blocked or undecided peers
    * `peers/` - Contains virtual directories for sending files.
        * `<peer_name>/` - Example peer directory
            * `<file_name>-<status>` - The sending status file contains details about progress, finish or error. 
