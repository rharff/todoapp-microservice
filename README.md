# TaskFlow Microservices

Local microservices playground: Task Service, Audit Service, and SvelteKit frontend.

## Services

- Task Service: `http://localhost:8080`
- Audit Service: `http://localhost:8081`
- Frontend: `http://localhost:5173`
- CloudBeaver: `http://localhost:8978`

## Local Docker (Fedora)

```bash
cd /home/rahan/Testing/Apps/todoapp-microservice
cp .env.example .env

docker compose up --build
```

## Frontend

```bash
cd /home/rahan/Testing/Apps/todoapp-microservice/frontend
cp .env.example .env
npm install
npm run dev
```

## Services (local without Docker)

Task Service:

```bash
cd /home/rahan/Testing/Apps/todoapp-microservice/services/task-service
cp .env.example .env
export $(cat .env | xargs)
go run ./cmd/server
```

Audit Service:

```bash
cd /home/rahan/Testing/Apps/todoapp-microservice/services/audit-service
cp .env.example .env
export $(cat .env | xargs)
go run ./cmd/server
```
