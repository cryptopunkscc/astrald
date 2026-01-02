# channel

A channel is a tool that allows you to encode and/or decode objects over a
bytestream using a variety of formats. Channels can be used for real-time
communication or to stream data to/from a file.

<!-- TOC -->
* [channel](#channel)
  * [Basic usage](#basic-usage)
    * [Creating a channel](#creating-a-channel)
    * [Sending to a channel](#sending-to-a-channel)
    * [Receiving from a channel](#receiving-from-a-channel)
    * [Closing a channel](#closing-a-channel)
  * [One-way channels](#one-way-channels)
  * [Asymmetric channels](#asymmetric-channels)
  * [Supported formats](#supported-formats)
    * [Binary](#binary)
    * [JSON](#json)
    * [Text](#text)
    * [Render](#render)
<!-- TOC -->

## Basic usage

### Creating a channel

A channel can be created over any `io.ReadWriter` using the `New` function:

```go
var transport io.ReadWriter // a net.Conn, a File, etc.
ch := channel.New(transport) // create a new channel over the transport
```

By default, the channel will use the binary format for encoding/decoding
objects. You can change this by passing config options to the `New` function:

```go
ch := channel.New(transport, channel.WithFormat("json"))
```

Channels can use different formats for reading and writing objects.

### Sending to a channel

Sending to a channel is done using the `Send` method:

```go
var obj astral.Object
err := ch.Send(obj)
```

The object needs to be fully marshalable to the output format of the channel.

### Receiving from a channel

Receiving from a channel is done using the `Receive` method:

```go
obj, err := ch.Receive()

switch obj := obj.(type) {
case *astral.Identity:
    // handle identity object
case *astral.ErrorMessage:
    // handle error message 
case nil:
	// receive error (err != nil)
default:
	// unknown object type
}
```

### Closing a channel

A channel can be closed using the `Close` method:

```go
err := ch.Close()
```

If the underlying transport is an io.Closer, the `Close` method will be called.
Otherwise, ErrCloseUnsupported will be returned.

## One-way channels

Sometimes you only want to send objects to a channel without receiving any
response. In this case, you can use a one-way channel:

```go
f := os.Create("file.bin")
ch := channel.NewSender(f)
ch.Send(obj)
```

And similarly for receiving:

```go
f := os.Open("file.bin")
ch := channel.NewReceiver(f)
for {
    obj, err := ch.Receive()
	if err != nil {
		break
	}
	// handle object
}
```

You can use the same config options as for regular channels, irrelevant options
will be ignored.

## Asymmetric channels

If you need to combine an independent io.Reader and an io.Writer into a
single channel, you can use the Join function:

```go
ch := channel.New(channel.Join(os.Stdin, os.Stdout))
```

The joined channel implements io.Closer by trying to close the io.Writer first,
then the io.Reader and finally returning ErrCloseUnsupported if neither of them
implements io.Closer.

## Supported formats

### Binary

This is the default format that's mandatory for all astral objects. It's
fast and efficient, but it's not human-readable. It's designed for machine-to-
machine communication. The objects in the channel are encoded as a String8
(representing the object type) followed by Byte32 buffer containing the object's
payload.

### JSON

This format is optional but widely supported. It's human-readable and
integrates well with existing technologies. Designed for communication with
front-end apps and debugging.

### Text

This format is optional and only supported by fairly simple data types that
can be serialized into a single line text snippet. It's human-readable and easy
to parse by hand. Used primarily to serialize objects as query arguments similar
to URL query parameters (e.g. `?key=value`).

### Render

This is a special output-only format that renders the object as a formatted
text. This is only used to present objects to the user.