# Chatyx - System Design

This page describes the design of message system.

## Requirements

The system should meet the following requirements:

### Functional requirements

- Support groups and dialogs chats
- Add and remove participants for group chats
- Participants can leave from group chats
- Block opponents in dialogs
- Send text messages and images
- View unread messages
- Show online/offline statuses of users, as well as when the user was last online
- If the user is not online, they should receive a notification
- Support mobile and web version
- Support cross-device synchronization
- Geo distribution is not supported (CIS only)
- No seasonality

### Non-functional requirements

- 100 000 000 DAU
- Availability 99.95% (4.38 hours downtime per year)
- The system should be scalable and efficient
- A group can have a maximum of 100 participants
- A message should reach the recipient in 3 seconds
- A message should be sent in 1 second
- Each user sends an average of 10 messages per day
- Each user reads messages an average of 20 times per day
- The size of each message is a maximum of 2000 characters
- The size of each image is a maximum 1 MB
- Each message has a maximum 3 images
- 5% of messages contain images
