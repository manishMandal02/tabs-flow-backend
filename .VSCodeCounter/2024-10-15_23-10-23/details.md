# Details

Date : 2024-10-15 23:10:23

Directory /Users/manish/main/programming/projects/tabs-flow/backend

Total : 69 files,  7841 codes, 291 comments, 1970 blanks, all 10102 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [.github/workflows/dev-deploy.yml](/.github/workflows/dev-deploy.yml) | YAML | 29 | 2 | 8 | 39 |
| [Makefile](/Makefile) | Makefile | 30 | 9 | 14 | 53 |
| [README.md](/README.md) | Markdown | 172 | 0 | 56 | 228 |
| [cmd/auth/lambda_authorizer/main.go](/cmd/auth/lambda_authorizer/main.go) | Go | 10 | 1 | 4 | 15 |
| [cmd/auth/main.go](/cmd/auth/main.go) | Go | 13 | 1 | 7 | 21 |
| [cmd/email/main.go](/cmd/email/main.go) | Go | 10 | 1 | 5 | 16 |
| [cmd/local_server/main.go](/cmd/local_server/main.go) | Go | 61 | 5 | 25 | 91 |
| [cmd/notes/main.go](/cmd/notes/main.go) | Go | 13 | 1 | 8 | 22 |
| [cmd/notifications/main.go](/cmd/notifications/main.go) | Go | 34 | 8 | 17 | 59 |
| [cmd/spaces/main.go](/cmd/spaces/main.go) | Go | 13 | 1 | 8 | 22 |
| [cmd/users/main.go](/cmd/users/main.go) | Go | 13 | 1 | 7 | 21 |
| [config/config.go](/config/config.go) | Go | 79 | 3 | 17 | 99 |
| [go.mod](/go.mod) | Go Module File | 47 | 0 | 4 | 51 |
| [go.sum](/go.sum) | Go Checksum File | 121 | 0 | 1 | 122 |
| [infra/README.md](/infra/README.md) | Markdown | 10 | 0 | 5 | 15 |
| [infra/bin/infra.ts](/infra/bin/infra.ts) | TypeScript | 24 | 2 | 5 | 31 |
| [infra/cdk.json](/infra/cdk.json) | JSON | 50 | 0 | 1 | 51 |
| [infra/config.ts](/infra/config.ts) | TypeScript | 53 | 2 | 11 | 66 |
| [infra/event.json](/infra/event.json) | JSON | 4 | 0 | 1 | 5 |
| [infra/jest.config.js](/infra/jest.config.js) | JavaScript | 8 | 0 | 1 | 9 |
| [infra/lib/stacks/services/auth.ts](/infra/lib/stacks/services/auth.ts) | TypeScript | 65 | 4 | 16 | 85 |
| [infra/lib/stacks/services/email.ts](/infra/lib/stacks/services/email.ts) | TypeScript | 49 | 3 | 13 | 65 |
| [infra/lib/stacks/services/index.ts](/infra/lib/stacks/services/index.ts) | TypeScript | 1 | 0 | 1 | 2 |
| [infra/lib/stacks/services/notes.ts](/infra/lib/stacks/services/notes.ts) | TypeScript | 43 | 2 | 8 | 53 |
| [infra/lib/stacks/services/notifications.ts](/infra/lib/stacks/services/notifications.ts) | TypeScript | 72 | 3 | 16 | 91 |
| [infra/lib/stacks/services/rest-api.ts](/infra/lib/stacks/services/rest-api.ts) | TypeScript | 40 | 12 | 11 | 63 |
| [infra/lib/stacks/services/service-stack.ts](/infra/lib/stacks/services/service-stack.ts) | TypeScript | 80 | 2 | 17 | 99 |
| [infra/lib/stacks/services/spaces.ts](/infra/lib/stacks/services/spaces.ts) | TypeScript | 40 | 2 | 8 | 50 |
| [infra/lib/stacks/services/users.ts](/infra/lib/stacks/services/users.ts) | TypeScript | 41 | 2 | 9 | 52 |
| [infra/lib/stacks/stateful/index.ts](/infra/lib/stacks/stateful/index.ts) | TypeScript | 1 | 0 | 1 | 2 |
| [infra/lib/stacks/stateful/stateful.ts](/infra/lib/stacks/stateful/stateful.ts) | TypeScript | 66 | 1 | 11 | 78 |
| [infra/package.json](/infra/package.json) | JSON | 29 | 0 | 1 | 30 |
| [infra/pnpm-lock.yaml](/infra/pnpm-lock.yaml) | YAML | 2,324 | 0 | 333 | 2,657 |
| [infra/test/infrastructure.test.ts](/infra/test/infrastructure.test.ts) | TypeScript | 2 | 13 | 3 | 18 |
| [infra/tsconfig.json](/infra/tsconfig.json) | JSON with Comments | 31 | 0 | 1 | 32 |
| [internal/auth/auth.go](/internal/auth/auth.go) | Go | 51 | 0 | 6 | 57 |
| [internal/auth/handlers.go](/internal/auth/handlers.go) | Go | 210 | 8 | 73 | 291 |
| [internal/auth/helpers.go](/internal/auth/helpers.go) | Go | 167 | 13 | 42 | 222 |
| [internal/auth/repository.go](/internal/auth/repository.go) | Go | 198 | 7 | 60 | 265 |
| [internal/auth/routes.go](/internal/auth/routes.go) | Go | 25 | 3 | 19 | 47 |
| [internal/email/email.go](/internal/email/email.go) | Go | 74 | 3 | 34 | 111 |
| [internal/email/zepto_mail.go](/internal/email/zepto_mail.go) | Go | 136 | 1 | 29 | 166 |
| [internal/notes/handlers.go](/internal/notes/handlers.go) | Go | 294 | 19 | 104 | 417 |
| [internal/notes/notes.go](/internal/notes/notes.go) | Go | 38 | 0 | 6 | 44 |
| [internal/notes/reposiotry.go](/internal/notes/reposiotry.go) | Go | 290 | 6 | 89 | 385 |
| [internal/notes/routes.go](/internal/notes/routes.go) | Go | 22 | 4 | 10 | 36 |
| [internal/notifications/events.go](/internal/notifications/events.go) | Go | 96 | 5 | 34 | 135 |
| [internal/notifications/handlers.go](/internal/notifications/handlers.go) | Go | 64 | 20 | 26 | 110 |
| [internal/notifications/notifications.go](/internal/notifications/notifications.go) | Go | 36 | 0 | 6 | 42 |
| [internal/notifications/repository.go](/internal/notifications/repository.go) | Go | 122 | 0 | 37 | 159 |
| [internal/notifications/routes.go](/internal/notifications/routes.go) | Go | 18 | 3 | 9 | 30 |
| [internal/spaces/handlers.go](/internal/spaces/handlers.go) | Go | 281 | 7 | 108 | 396 |
| [internal/spaces/repository.go](/internal/spaces/repository.go) | Go | 402 | 4 | 128 | 534 |
| [internal/spaces/routes.go](/internal/spaces/routes.go) | Go | 27 | 7 | 13 | 47 |
| [internal/spaces/spaces.go](/internal/spaces/spaces.go) | Go | 69 | 0 | 10 | 79 |
| [internal/users/handlers.go](/internal/users/handlers.go) | Go | 304 | 17 | 117 | 438 |
| [internal/users/helpers.go](/internal/users/helpers.go) | Go | 161 | 10 | 54 | 225 |
| [internal/users/repository.go](/internal/users/repository.go) | Go | 330 | 13 | 102 | 445 |
| [internal/users/routes.go](/internal/users/routes.go) | Go | 32 | 7 | 15 | 54 |
| [internal/users/users.go](/internal/users/users.go) | Go | 151 | 30 | 31 | 212 |
| [pkg/database/database.go](/pkg/database/database.go) | Go | 87 | 5 | 13 | 105 |
| [pkg/database/helpers.go](/pkg/database/helpers.go) | Go | 69 | 3 | 23 | 95 |
| [pkg/events/events.go](/pkg/events/events.go) | Go | 77 | 3 | 23 | 103 |
| [pkg/events/queue.go](/pkg/events/queue.go) | Go | 50 | 1 | 17 | 68 |
| [pkg/events/scheduler.go](/pkg/events/scheduler.go) | Go | 54 | 6 | 14 | 74 |
| [pkg/http_api/http_api.go](/pkg/http_api/http_api.go) | Go | 48 | 0 | 9 | 57 |
| [pkg/http_api/router.go](/pkg/http_api/router.go) | Go | 98 | 4 | 28 | 130 |
| [pkg/logger/logger.go](/pkg/logger/logger.go) | Go | 17 | 0 | 9 | 26 |
| [pkg/utils/utils.go](/pkg/utils/utils.go) | Go | 65 | 1 | 18 | 84 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)