# Product

This document is a sketchbook for noting thoughts about product.

# Description

The core feature of the application is a file sending, so the user can act with the application as a sender or a
recipient. The sender is sending selected files to the recipient. The recipient is notified about offered files and can
decide if he wants to download files or not. Both sides can track the file transfer progress and are notified about
finish. Additionally, the user can mark a peer as blocked or trusted so the application will automatically reject file
offers or download the files.

# Epic 1

> As a sender, I can send files offer to my mate, but as a recipient, I can accept files offer from my mate.

## Use cases

> The sender is sending files offer to the recipient.

1. User chooses the files to offer;
2. User chooses a recipient from connected peers;
3. Application is sending the files offer to selected peer;
4. User receives a feedback about;
   * Success - standard info that files offer was successfully sent.
   * Error - short description of specific error.

> The recipient is notified about a files offer.

1. Application receives files offer;
2. Application saves offer in storage;
3. Application notifies the user about files offer;
4. User sees files offer notification;
5. User navigates from notification to offer details;
6. User sees files offer details;

> The recipient is accepting files offer from sender so downloading is started.

1. User sees the files offer.
2. User is accepting files offer.
3. Application is downloading files in background.

> The user is tracking file transferring status.

1. Application is downloading files.
2. User tracks transfer progress status.

> The user is notified about transfer finish.

1. Application finishes file downloading because:
   * is completed with success;
   * the error occurs;
2. Application notifies user about download finish;
3. User sees files transfer finish notification.

# Epic 2

> I want the app will automatically download the files offered by trusted mate.

> I can block the mate, to prevent me from receiving files from him.

