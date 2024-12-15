# TabsFlow Backend

> AWS Serverless architecture with **Golang** for core services and **TypeScript** for CDK infrastructure code.

## Table of Contents

- [Project Overview](#project-overview)

- [Technologies](#technologies)

- [Architecture](#architecture)

- [Entities](#entities)

- [Data Access Patterns](#data-access-patterns-main-table)

- [Table Designs (DynamoDB)](#main-table-design-dynamodb)

- [Data Types](#data-types)

- [Services](#services)

- [Folder Structure](#folder-structure)

- [Getting Started](#getting-started)

- [Deployment](#deployment)

## Project Overview

TabsFlow Backend is a serverless application built on AWS, designed to manage tabs, spaces, and user data. It provides a robust backend infrastructure for the TabsFlow web application.

## Technologies

- Programming Language: Golang (for services) | TypeScript for CDK infra

- Cloud Provider: AWS

  - CDK (Infrastructure as Code)

  - API Gateway (REST APIs)

  - Lambda (Compute Services)

  - DynamoDB (Main Data Store)

  - SQS (Asynchronous Tasks)

  - CloudWatch (Monitoring & Logging)

  - EventBridge Scheduler (Scheduling Tasks)

- Transactional Emails Service: Zoho ZeptoMail

## Architecture

[TODO: high-level architecture diagram here]

## Entities

| Entity                   | Description                    | Attributes                                                    |
| ------------------------ | ------------------------------ | ------------------------------------------------------------- |
| User                     | user profile                   | Id, email, FullName, ProfilePic,                              |
| Preferences              | app preferences                | UserId, CmdPalette{}, Notes{}, AutoDiscard{}, OpenSpace, etc. |
| Space                    | workspaces                     | Id, UserId, Title, Emoji, Theme, WindowId ActiveTabIndex      |
| Tab                      | tabs within space              | SpaceId, Index, Title, URL, FaviconURL, GroupId               |
| Group                    | tab groups                     | Id, SpaceId, Title, Color, Collapsed                          |
| Note                     | user notes                     | Id, UserId, SpaceId,, Title, Note, RemainderAt, UpdatedAt     |
| SnoozedTab               | snoozed tabs in space          | SpaceId, Title, URL, FaviconURL, SnoozedUntil                 |
| Notification             | notifications & remainders     | UserId, Type, Timestamp, Note{}, SnoozedTab{}                 |
| NotificationSubscription | notification subscription info | UserId,Endpoint, AuthKey, P256dhKey                           |
| Subscription             | user subscriptions             | Id, PlanId, Plan, Status, Start, End, NextBillingDate         |

## Data Access Patterns (Main Table)

| Access Pattern              | Entities Retrieved |
| --------------------------- | ------------------ |
| Get User by id              | User               |
| Get Preferences by userId   | Preferences        |
| Get Spaces by userId        | Spaces             |
| Get Tabs by spaceId         | Tabs               |
| Get Groups by spaceId       | Tabs               |
| Get SnoozedTabs by spaceId  | SnoozedTabs        |
| Get Notes by userId         | Notes              |
| Get Notifications by userId | Notifications      |
| Get Subscription by userId  | Subscription       |

## Main Table Design (DynamoDB)

| Partition Key (PK) | Sort Key (SK)                          | Item Attributes                                                     |
| ------------------ | -------------------------------------- | ------------------------------------------------------------------- |
| {UserId}           | U#Profile                              | Email, FullName, ProfilePic, UpdatedAt                              |
|                    | U#Subscription                         | Id, PlanId, Plan, Status, Start, End, NextBillingDate               |
|                    | P#General                              | OpenSpace, DeleteUnsavedSpace, UpdatedAt                            |
|                    | P#Notes                                | IsDisabled, BubblePos, ShowOnAllSites, UpdatedAt                    |
|                    | P#CmdPalette                           | IsDisabled, Search, DisabledCommands, UpdatedAt                     |
|                    | P#LinkPreview                          | IsDisabled, OpenTrigger, Size, UpdatedAt                            |
|                    | P#AutoDiscard                          | IsDisabled, DiscardAfter, WhitelistedDomains, UpdatedAt             |
|                    | U#Notification#{Id/CreatedAt}          | Type, Timestamp, Note{}, SnoozedTab{}                               |
|                    | U#NotificationSubscription             | UserId,Endpoint, AuthKey, P256dhKey                                 |
|                    | S#Info#{SpaceId}                       | Title, Emoji, Theme, ActiveTab, windowId, ActiveTabIndex, UpdatedAt |
|                    | S#Tabs#{SpaceId}                       | []{ Index, Title, URL, FaviconURL, GroupId }, UpdatedAt             |
|                    | S#Groups#{SpaceId}                     | []{ Title, Color, Collapsed }, UpdatedAt                            |
|                    | S#SnoozedTabs#{SpaceId}#{Id/SnoozedAt} | SpaceId, Title, URL, FaviconURL, SnoozedUntil, SnoozedAt            |
|                    | N#{NoteId/CreatedAt}                   | Id, SpaceId, Title, Note, RemainderAt, UpdatedAt                    |

## Data Access Patterns (Search Table)

| Access Pattern              | Attributes Retrieved |
| --------------------------- | -------------------- |
| Get NoteIds by search terms | NoteId               |

## Notes Search Table Design (DynamoDB)

| Partition Key (PK) | Sort Key (SK) | Item Attributes                                  |
| ------------------ | ------------- | ------------------------------------------------ |
| {UserId}#{Term}    | N#NoteId      | Id, SpaceId, Title, Note, RemainderAt, UpdatedAt |

## Data Access Patterns (Sessions Table)

| Access Pattern | Attributes Retrieved |
| -------------- | -------------------- |
| User by email  | UserID               |
| Check OTP      | OTP                  |
| Get Session    | Session{}            |

## Sessions Table Design (DynamoDB)

| Partition Key (PK) | Sort Key (SK)   | Item Attributes            |
| ------------------ | --------------- | -------------------------- |
| {EmailId}          | UserId#{userId} |                            |
|                    | OTP#{otp}       | TTL                        |
|                    |                 |                            |
| {UserId}           | S#{sessionId}   | CreatedAt, DeviceInfo, TTL |

## Data Types

- Date & time

- All dates & times are stored in UTC Unix timestamps (seconds since epoch).

- Pricing Plans: TRAIL | YEARLY | LIFETIME

- Subscription Status: active | canceled | past_due | paused | trialing

- API Response

- Success

- Status Code: 200 (OK)

- Response Body:

```json

{

"success": true,

"message": "success message", // only for POST/PATCH requests

"data": JSON  object  of  the  requested  Entity, ex: User, []notes,  note,  []notifications  etc.

}

```

- Error

- Status Code: 500s (server error) or 400s (bad request)

- Response Body: (some endpoints don't response with json body, just status code and the error message text)

```json
{ "success": false, "message": "error message" }
```

## Services

### Auth Service

- Handles authn & authz: Google OAuth, OPT verification, and session management.

- API Endpoints: /auth

- Get: /verify-otp

- Get: /user-id

- Get: /logout

- POST: /google

- POST: /send-otp

- Env variables:

- JWT_SECRET_KEY

- EMAIL_QUEUE_URL

- DDB_SESSIONS_TABLE_NAME

### Users Service

- Manages user-related operations

- API Endpoints: /users

- POST: /

- GET: /:id

- PATCH: /:id

- DELETE: /:id

- GET: /:id/preferences

- PATCH: /:id/preferences

- GET: /:id/subscription

- GET: /:id/subscription/status

- GET: /:id/subscription/paddle-url

- POST: /:id/subscription/webhook

- Env variables:

- EMAIL_QUEUE_URL

- DDB_MAIN_TABLE_NAME

- PADDLE_API_KEY

- PADDLE_WEBHOOK_SECRET_KEY

### Spaces Service

- Manages spaces, tabs, and groups

- API Endpoints: /spaces

- POST: /:userId

- GET: /:spaceId

- GET: /user/:userId

- PATCH: /:spaceId

- DELETE: /:spaceId

- Env variables:

- DDB_MAIN_TABLE_NAME

### Notes Service

- Manages spaces, tabs, and groups

- API Endpoints: /spaces

- POST: /:userId

- GET: /:userId/:noteId

- GET: /:userId (notes by user)

- PATCH: /:userId

- DELETE: /:userId/:noteId

- Env variables:

- DDB_MAIN_TABLE_NAME

- DDB_SEARCH_INDEX_TABLE_NAME

### Notifications Service

- Handles notification management

- API Endpoints: /notifications

- GET: /{userId}

- POST: /{userId}

- DELETE: /{userId}

### Email Service

- Sends transactional emails

- No direct API access, polls an SQS queue for messages

- Events:

- SEND_OTP

- USER_REGISTERED

- Env variables:

- ZEPTO_MAIL_API_KEY

- EMAIL_QUEUE_URL

### Scheduler Service

- Schedules tasks for future execution

- No direct API access

- Polls an SQS queue for messages to schedule tasks (e.g., note reminders)

### Monitoring Service

- Handles monitoring and observability

- Services send metrics to CloudWatch

- Sets up alerts/alarms for metrics

- Implements structured logging (sent to CloudWatch)

## Folder Structure

- `/cmd` - main entry points for each service

- `/internal` - services code domain, routes, handlers, repository

- `/pkg` - shared code and utilities

- `/test` - unit tests

- `/infra` - infrastructure code (AWS CDK)

- `Makefile` - dev, build and test commands

- `README.md` - project documentation

- `.air.toml` - Air configuration file for live reloading Go files

- `.env.example` - example environment variables

## Getting Started

1. Clone the repository and navigate to the project directory.

```bash

git clone https://github.com/manishMandal02/freshtabs-backend.git

```

2. Add environment variables: (check .env.example)

3. Install dependencies:

```bash

make init

```

4. Run the project:

```bash

make dev

```

## Deployment

[TODO: Instructions on how to deploy the project to AWS]
