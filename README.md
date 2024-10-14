# TabsFlow Backend

> AWS Serverless architecture with **Golang** for core services and **TypeScript** for CDK infrastructure code.

## Table of Contents

- [Project Overview](#project-overview)
- [Technologies](#technologies)
- [Architecture](#architecture)
- [Entities](#entities)
- [Data Access Patterns](#data-access-patterns-dynamodb)
- [Table Design (DynamoDB)](#table-design-dynamodb)
- [Services](#services)
- [Getting Started](#getting-started)
- [Development](#development)
- [Deployment](#deployment)

## Project Overview

TabsFlow Backend is a serverless application built on AWS, designed to manage tabs, spaces, and user data. It provides a robust backend infrastructure for the TabsFlow web application.

## Technologies

- **Core Language**: Golang (for services)
- **Infrastructure as Code**: TypeScript (for AWS CDK)
- **Cloud Provider**: AWS
  - CDK (Infrastructure as Code)
  - API Gateway (REST APIs)
  - Lambda (Compute Services)
  - DynamoDB (Main Data Store)
  - SQS (Asynchronous Tasks)
  - CloudWatch (Monitoring & Logging)
  - EventBridge (Scheduling Future Tasks)
- **Zoho ZeptoMail**: Transactional Emails

## Architecture

[TODO: high-level architecture diagram here]

Our serverless architecture leverages various AWS services to create a scalable and maintainable backend system. The core services are implemented as Lambda functions, with API Gateway serving as the main entry point for client requests.

## Entities

| Entity       | Description               | Attributes                                                    |
| ------------ | ------------------------- | ------------------------------------------------------------- |
| User         | user profile              | Id, email, FullName, ProfilePic,                              |
| Preferences  | app preferences           | UserId, CmdPalette{}, Notes{}, AutoDiscard{}, OpenSpace, etc. |
| Space        | workspaces                | Id, UserId, Title, Emoji, Theme, WindowId ActiveTabIndex      |
| Tab          | tabs within space         | SpaceId, Index, Title, URL, FaviconURL, GroupId               |
| Group        | tab groups                | Id, SpaceId, Title, Color, Collapsed                          |
| Note         | user notes                | Id, UserId, SpaceId,, Title, Note, RemainderAt, UpdatedAt     |
| SnoozedTab   | snoozed tabs in space     | SpaceId, Title, URL, FaviconURL, SnoozedUntil                 |
| Notification | notification & remainders | UserId, Type, Timestamp, Note{}, SnoozedTab{}                 |
| Subscription | user subscriptions        | Id, PlanId, Plan, Status, Start, End, NextBillingDate         |

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

| Partition Key (PK) | Sort Key (SK)                          | Item Attributes                                          |
| ------------------ | -------------------------------------- | -------------------------------------------------------- |
| UserId             | U#Profile                              | Email, FullName, ProfilePic                              |
|                    | U#Subscription                         | Id, PlanId, Plan, Status, Start, End, NextBillingDate    |
|                    | P#General                              | IsDisabled, DiscardAfter, WhitelistedDomains             |
|                    | P#Note                                 | IsDisabled, BubblePos, ShowOnAllSites                    |
|                    | P#CmdPalette                           | IsDisabled, Search, DisabledCommands                     |
|                    | P#LinkPreview                          | IsDisabled, OpenTrigger, Size                            |
|                    | P#AutoDiscard                          | IsDisabled, DiscardAfter, WhitelistedDomains             |
|                    | U#Notification#{Id/CreatedAt}          | Type, Timestamp, Note{}, SnoozedTab{}                    |
|                    | S#Info#{SpaceId}                       | Title, Emoji, Theme, ActiveTab, windowId, ActiveTabIndex |
|                    | S#Tabs#{SpaceId}                       | []{ Index, Title, URL, FaviconURL, GroupId }             |
|                    | S#Groups#{SpaceId}                     | []{ Title, Color, Collapsed }                            |
|                    | S#SnoozedTabs#{SpaceId}#{Id/SnoozedAt} | SpaceId, Title, URL, FaviconURL, SnoozedUntil, SnoozedAt |
|                    | N#{NoteId/CreatedAt}                   | Id, SpaceId, Title, Note, RemainderAt, UpdatedAt         |

## Data Access Patterns (Search Table)

| Access Pattern       | Attributes Retrieved |
| -------------------- | -------------------- |
| Get NoteIds by query | NoteId               |

## Notes Search Table Design (DynamoDB)

| Partition Key (PK) | Sort Key (SK) | Item Attributes                                  |
| ------------------ | ------------- | ------------------------------------------------ |
| UserId#Term        | N#NoteId      | Id, SpaceId, Title, Note, RemainderAt, UpdatedAt |

## Data Access Patterns (Sessions Table)

| Access Pattern | Attributes Retrieved |
| -------------- | -------------------- |
| User by email  | UserID               |
| Check OTP      | OTP                  |
| Get Session    | Session{}            |

## Sessions Table Design (DynamoDB)

| Partition Key (PK) | Sort Key (SK)       | Item Attributes            |
| ------------------ | ------------------- | -------------------------- |
| EmailId            | UserId#{userId}     |                            |
|                    | OTP#{otp}           | TTL                        |
|                    | Session#{sessionId} | CreatedAt, TTL, DeviceInfo |

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

## Getting Started

[TODO: Instructions on how to set up the project locally]

## Development

[TODO: Guidelines working on the project locally]

## Deployment

[TODO: Instructions on how to deploy the project to AWS]
