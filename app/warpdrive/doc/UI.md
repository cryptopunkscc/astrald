# UI

Abstract point of view on how the interaction with warpdrive app could look. 

## `sender`

### `send`

1. Choose a file to send;
2. Choose a recipient from connected peers;
3. Receive a feedback:
    1. Error.
    2. Sending status:
        1. Waiting for send;
        2. Sending progress;
        3. Sending completed with:
            1. Success.
            2. Error.

## `recipient`

### `incoming`

1. Observe new notification about incoming file request.

### `accept`

1. Choose `accept` option for incoming file request.

### `reject`

1. Choose `reject` option for incoming file request.

### `update`

1. Open peers list
2. Choose a proper option for required peer:
    1. Ask - default option.
    2. Trust - accept incoming files without asking.
    3. Block - block any incoming file request.
    