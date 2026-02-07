# 📋 TaskFlow: Microservices Vibecoding Blueprint

> **System Prompt for Agent:** Act as a Senior Full-Stack Engineer. Follow the architectural boundaries strictly. Use Go standard library + `chi` for routing, `pgx` for Postgres, and SvelteKit for the frontend.

---

## 🏗️ 1. Global Project Context

* **Architecture:** Two isolated microservices + one frontend.
* **Communication:** Task Service calls Audit Service via REST.
* **Constraint:** Services MUST NOT share a database. Each has its own PostgreSQL instance/database.
* **Language Stack:** Go (Backend), Svelte + Tailwind (Frontend), PostgreSQL (Database).

---

## 🛠️ 2. Folder Structure

```text
/taskflow
  ├── services/
  │   ├── task-service/     # Port: 8080 | DB: task_db
  │   └── audit-service/    # Port: 8081 | DB: audit_db
  ├── frontend/             # Port: 5173 (SvelteKit + Tailwind)
  ├── docker-compose.yml    # Infrastructure
  └── PLAN.md               # This file

```

---

## 🎯 3. Phase-by-Phase Achievement Goals

### Phase 1: Infrastructure & Database (`docker-compose.yml`)

* **Goal:** Spin up two separate Postgres containers and a `dbeaver` access point.
* **Verification:** `docker ps` shows two healthy DB instances.

### Phase 2: Audit Service (The "Passive" Service)

* **Goal:** Build an append-only logging API.
* **Requirements:**
* Endpoint: `POST /logs`.
* Payload: `{ "task_id": "uuid", "action": "string", "payload": "json" }`.
* Logic: Store in `audit_logs` table.


* **Tech:** Go + `pgx/v5`.

### Phase 3: Task Service (The "Active" Service)

* **Goal:** Build the CRUD core with an outbound trigger.
* **Requirements:**
* Endpoints: `GET /tasks`, `POST /tasks`, `PUT /tasks/:id`.
* **The Hook:** After a successful DB write, the Task Service must fire an asynchronous goroutine to `POST` the event to the Audit Service.


* **Tech:** Go + `chi` router + `pgx/v5`.

### Phase 4: Frontend (The "Vibe" Layer)

* **Goal:** A clean, reactive UI to manage tasks.
* **Requirements:**
* Task Dashboard: List tasks with "Complete" buttons.
* Audit Tab: A simple table showing the logs pulled directly from the Audit Service.


* **Tech:** SvelteKit + Tailwind CSS + Lucide Icons.

---

## 📐 4. Data Models (Reference)

### Task Service Schema

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, done
    created_at TIMESTAMP DEFAULT NOW()
);

```

### Audit Service Schema

```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    task_id UUID,
    action_string TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

```

---

## 🚦 5. Implementation Rules for AI Agent

1. **Strict Typing:** Generate Go structs that match the SQL schema exactly.
2. **Error Handling:** Don't just `log.Fatal`. Return proper JSON error responses (`{ "error": "message" }`).
3. **CORS:** Enable CORS on both Go services so the Svelte frontend (running on a different port) can talk to them.
4. **Environment Variables:** Use `.env` files for `DATABASE_URL` and `AUDIT_SERVICE_URL`.
