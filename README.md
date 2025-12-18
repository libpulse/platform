# LibPulse Platform

**The platform powering LibPulse — an open-source observability system for devtools and developer-facing infrastructure.**

LibPulse helps devtool authors understand how their tools are used in the real world — without turning analytics into a nightmare.

> 🚧 **Status: Active development**
>
> LibPulse is built in public. APIs are evolving, features are incomplete, and breaking changes may happen.
>
> Feedback, early users, and contributors are very welcome.

---

## What is LibPulse?

LibPulse is an observability platform designed for developer tools.

It focuses on **retention, debuggability, and real-world usage**, rather than one-off demos or short-lived apps.

This repository contains the **core LibPulse platform**, including both backend and frontend components:

- **Backend APIs** written in Go, designed for high-throughput ingestion and clear data boundaries.
- **Data storage and processing** powered by Supabase, with a strong focus on multi-tenant data modeling and access control.
- **Frontend dashboard** built with React and TypeScript, used to explore and debug real telemetry data.
- **Observability & visualization layer** powered by Grafana *(work in progress)*.

👉 **Language SDKs are developed in separate repositories**.  
This repository intentionally focuses on the **platform core**, while SDKs evolve independently per language and ecosystem.

## Why LibPulse?

If you're building a devtool (CLI, SDK, or infrastructure product), you probably want to know:

- Which features are actually used?
- Where users get stuck?
- Which versions introduce breaking behavior?
- How real developers interact with your tool?

Most analytics platforms are built for **products and marketing**, not for **developer tools**.
They are often too heavy, too opaque, or simply not designed for infra-level insights.

**LibPulse is built specifically for devtools.**


## Local Development (for contributors)

LibPulse is under active development.

This section is intended for developers who want to explore the codebase or contribute to the platform. 

LibPulse relies on Supabase schema and RLS policies. If you create a new Supabase project for local development, you will usually need Supabase CLI once to apply the schema.
After that, Supabase CLI is not required for day-to-day development, unless your PR changes database schema or RLS.

Contributors are expected to create their own Supabase project (free tier is sufficient) and use their own keys locally.


### Frontend Prerequisites
	1. Node.js ≥ 22.9.0
	2. pnpm installed

For environment variables, you need to a create a file `web/.env.local`:

```shell
// web/.env.local
NEXT_PUBLIC_SUPABASE_URL=https://<project-ref>.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=XXX
```

You can find these values in the Supabase Dashboard:
	•	NEXT_PUBLIC_SUPABASE_URL:  Project Settings → Data API → Project URL
	•	NEXT_PUBLIC_SUPABASE_ANON_KEY:  Project Settings → API Keys → Legacy anon, service_role API keys → anon
> Noted: These are public client-side keys. Never use service_role key in the frontend.


### Backend Prerequisites
	Go ≥ 1.25.5

For environment variables, you need to a create a `.env.dev` file in the repository root:
```shell
// .env.dev
SUPABASE_JWT_SECRET=XXX
SUPABASE_SERVICE_ROLE_KEY=XXX
SUPABASE_AUTH_URL=XXX
```
Also, in `.env.dev` file, you need to set your CORS origins, like this:
```shell
CORS_ALLOWED_ORIGINS=http://localhost:3000
```
Where to find these values:
	•	SUPABASE_SERVICE_ROLE_KEY:  Project Settings → API Keys → Legacy anon, service_role API keys → service_role
	•	SUPABASE_JWT_SECRET: Project Settings → JWT Keys → Legacy JWT Secret
	•	SUPABASE_AUTH_URL: ${SUPABASE_PROJECT_URL}/auth/v1

> NOTED: SUPABASE_SERVICE_ROLE_KEY is sensitive. Keep it in .env.dev only and never commit it.


### Apply schema to your Supabase project

The backend expects the database schema and RLS policies to already exist. If your Supabase project is newly created, you must apply the schema before running the backend.

Steps:
	1.	Install Supabase CLI
	2.	Run the following commands in the repository root:
```shell
supabase login
supabase link
supabase db push
```

This will apply all schema and RLS policies defined in `supabase/migrations`.

Migration files are the single source of truth. Please do not modify schema directly via Supabase Dashboard.

After this step, Supabase CLI is not required unless your PR modifies database schema or RLS.

### Run the project

Once the prerequisites and schema are ready, start the project:

```shell
git clone https://github.com/libpulse/platform.git
cd platform
```

#### install dependencies
In the root directory, you can execute this command to install dependencies:
```shell
make setup
```

#### run frontend + backend
Before running, make sure:
	•	`web/.env.local` exists
	•	`.env.dev` exists
	•	Supabase schema has been applied
  
In the root directory, you can execute this command to run the platform(frontend + backend):
```shell
make dev
```
After you run `make dev`, you will see:

•	Frontend runs at: http://localhost:3000
•	Backend runs at: http://localhost:8080


#### Run tests

Run unit tests (handlers only, fast):
```shell
make test
```

Run tests with verbose output and no cache:

```shell
make test-v
```