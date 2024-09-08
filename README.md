# TabsFlow Backend

> AWS Serverless architecture with **Golang** for core services and **TypeScript** for CDK infrastructure code.

## Table of Contents

- [Project Overview](#project-overview)
- [Technologies](#technologies)
- [Architecture](#architecture)
- [Entities](#entities)
- [Data Access Patterns](#data-access-patterns-dynamodb)
- [Services](#services)
- [Getting Started](#getting-started)
- [Development](#development)
- [Deployment](#deployment)

## Project Overview

TabsFlow Backend is a serverless application built on AWS, designed to manage tabs, spaces, and user data. It provides a robust backend infrastructure for [brief description of what TabsFlow does].

## Technologies

- **Core Language**: Golang (for services)
- **Infrastructure as Code**: TypeScript (for AWS CDK)
- **Cloud Provider**: AWS
  - CDK (Infrastructure as Code)
  - API Gateway (REST APIs)
  - Cognito (Authentication & Authorization)
  - Lambda (Compute Services)
  - DynamoDB (Main Data Store)
  - SQS (Asynchronous Tasks)
  - SES (Transactional Emails)
  - CloudWatch (Monitoring & Logging)
  - EventBridge (Scheduling Future Tasks)

## Architecture

[TODO - high-level architecture diagram here]

Our serverless architecture leverages various AWS services to create a scalable and maintainable backend system. The core services are implemented as Lambda functions, with API Gateway serving as the main entry point for client requests.

## Entities

| Entity         | Description               | Attributes                                                 |
| -------------- | ------------------------- | ---------------------------------------------------------- |
| User           | user profile              | Id, email, FullName, ProfilePic,                           |
| Preferences    | app preferences           | UserId, CmdPalette{}, Note{}, AutoDiscard{}, etc.          |
| Space          | workspaces                | Id, UserId, Title, Emoji, Theme, ActiveTab                 |
| Tab            | tabs within space         | SpaceId, Index, Title, URL, FaviconURL, GroupId            |
| Group          | tab groups                | Id, SpaceId, Title, Color, Collapsed                       |
| Note           | user notes                | Id, UserId, SpaceId,, Title, Note, RemainderAt, UpdatedAt  |
| SnoozedTab     | snoozed tabs in space     | SpaceId, Title, URL, FaviconURL, SnoozedUntil              |
| UsageAnalytics | space usage analytics     | UserId, SpaceUsage{}                                       |
| Notification   | notification & remainders | Id, UserId, Type, Timestamp, Note{}, SnoozedTab{}          |
| Subscription   | user subscriptions        | UserId, Plan, StartDate, EndState, Validity, TrailEndDate, |

## Data Access Patterns (DynamoDB)

| Access Pattern               | Entities Retrieved |
| ---------------------------- | ------------------ |
| Get User by id               | User               |
| Get User by email            | User               |
| Get Preferences by userId    | Preferences        |
| Get Spaces by userId         | Spaces             |
| Get Tabs by spaceId          | Tabs               |
| Get Groups by spaceId        | Tabs               |
| Get SnoozedTabs by spaceId   | SnoozedTabs        |
| Get Notes by spaceId         | Notes              |
| Get Notes by userId          | Notes              |
| Get Notifications by userId  | Notifications      |
| Get Subscription by userId   | Subscription       |
| Get UsageAnalytics by userId | UsageAnalytics     |

## Services

### Auth Service

- Handles authentication and authorization
- Features:
  - Google OAuth integration
  - API Gateway Authorizer
  - Email-based OTP verification
  - JWT session management

### User Service

- Manages user-related operations
- Endpoints:
  - GET/PUT: Profile
  - GET/PUT: Preferences
  - GET/PUT: Subscription

### Spaces Service

- Manages tabs, spaces, and groups
- Endpoints:
  - GET/PUT: Spaces
  - GET/PUT: Tabs
  - GET/PUT: Groups

### Sync Service

- Synchronizes local data with the cloud
- Endpoints:
  - POST: Sync data

### UsageAnalytics Service

- Tracks and manages space usage data
- Endpoints:
  - GET: Retrieve space usage data
  - PATCH: Update user space usage

### Notification Service

- Handles notification management
- Endpoints:
  - GET: Retrieve notifications
  - POST: Create notification
  - DELETE: Delete notification

### Email Service

- Sends transactional emails
- No direct API access
- Polls an SQS queue (Email Queue) for messages to send emails

### Scheduler Service

- Schedules tasks for future execution
- No direct API access
- Polls an SQS queue (Scheduler Queue) for messages to schedule tasks (e.g., note reminders)

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
