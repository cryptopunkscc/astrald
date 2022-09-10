# Technical Details

Warpdrive consist two applications:

* Background service
    * Same implementation independent of operating system.
    * Serves API for warpdrive UI client.
    * Communicates with UI client and other instances of warpdrive service.
* UI client
    * Can differ depending on OS.
    * Allows the user:
        * sending files to other warpdrive users.
        * receiving notifications.
        * downloading files.

## Architecture

```
client[sender] <=> service[1] <=> service[2] <=> client[recipient]
```

## OS requirements

The operating system allows the application for:

* running ongoing background service written in golang.
* inter-process communication, web or unix sockets.
* displaying and updating the notification.
* access the file from user space as byte-stream.
* saving files in user space.

# Persistent storage

* Sent offers
* Received offers
* Peers
* Files
