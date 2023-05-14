# Linux-Chat-App

## Summary

cli-client and server chat application implemented in Go

## Design Document

1. What kind of chat app will it be?

    it will support 1 on 1 and group chats

2. what is the group chat limit?

    100

3. Security?

    oppenssl certs

## High-level design

* services

```
+------------------+                   +-------------------+                     +--------------------+
|                  |                   |                   |                     |                    |
|                  |                   |                   |                     |                    |
|                  |                   |                   |                     |                    |
|                  |    Messsage       |                   |      Message        |                    |
|     Sender       +------------------->    Chat Server    +-------------------->|       Receiver     |
|                  |                   |                   |                     |                    |
|                  |                   |                   |                     |                    |
|                  |                   |                   |                     |                    |
|                  |                   |                   |                     |                    |
+------------------+                   +-------------------+                     +--------------------+

```
* data

```
     users                                                     chatRooms
+--------------------+                                    +--------------------+
|  array:            |                                    |   array:           |
|  +-------------+   |                                    |  +---------------+ |
|  | userName    |   |       chatRooms has chatroom       |  | chatRoomName  | |
|  | chatRooms   +---+----------------------------------->|  | users         | |
|  +-------------+   |      users has user                |  |               | |
|                    |<-----------------------------------+--+---------------+ |
|                    |                                    |                    |
|                    |                                    |                    |
|                    |                                    |                    |
|                    |                                    |                    |
|                    |                                    |                    |
+--------------------+                                    +--------------------+
```

* Messages Flow

```
+-----------------------+                         +---------------------+                 +-----------------------+
|                       |                         |                     |                 |                       |
|                       |                         |                     |                 |                       |
|     user              |                         |   chat service      |                 |     dataStore         |
|                       |      send message       |                     |  add message    |                       |
|                       +-------------------------+>                    +---------------->|                       |
|                       |                         |                     |                 |                       |
|                       |                         |                     |                 |                       |
|                       |                         |                     |                 |                       |
|                       |                         |                     |                 |                       |
+-----------------------+                         +---------------------+                 +-----------+-----------+
                                                                                                      |
                                                                                                      |
                                                                                                      |
                                                                                                      |
                                                                                                      |
                                                                                          +-----------v--------------+
                                                                                          |                          |
                                                                                          |                          |
                                                                                          |                          |
                                                                 sends out messages to users  messageQueue           |
                                                                                          |                          |
                                                               <--------------------------+                          |
                                                                                          |                          |
                                                                                          |                          |
                                                                                          |                          |
                                                                                          +--------------------------+
```

## Sample Run

![sampleRun](./sampleRun.gif)