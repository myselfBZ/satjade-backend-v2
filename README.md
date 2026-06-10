# Welcome to the backend of the SAT Jade

## This repo powers the https://www.satjade.live 's backend system.


Tech stack used:
    - Golang with Echo for HTTP
    - PostgreSQL
    - sqlc (to reduce quering redundancy and increase type safety)
    - Websockets for Duel competitive matches

Requirements to spin it up locally:
    - Go >=1.25.1
    - PostgreSQL
    - golang-migrate >=4.19

Run the following commands in a sequential order to have the system up and running:
```bash
go mod tidy
```


```bash
migrate -path ./cmd/migrate -database "${YOUR_POSTGRES_URL}" up
```

In order to seed an administrator in the system run:
```bash
go run ./cmd/migrate/seed/admin
```

Start the API
```bash
go run ./cmd/api/
```

Optional, for alerting
```bash
go run ./cmd/alert/
```

env variables
```bash
#for the API binary
SECRET_KEY=
REFRESH_SECRET_KEY=
SERVER_PORT=
DB=
FRONTEND_URL=
TOKEN_EXPR_HOURS=
IMAGE_STORE=
LLM_APIKEY=
AUTH_AUD=
AUTH_ISS=
LOG_FILE=

#for the alert binary
TG_BOT_API_KEY=
CHAT_ID=

#for the admin seed binary
ADMIN_EMAIL=
ADMIN_PASSWORD=

```
