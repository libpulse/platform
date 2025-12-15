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

LibPulse is a full-stack observability platform designed for developer tools.

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

---


## Local Development (for contributors)

LibPulse is under active development.

This section is intended for developers who want to explore the codebase or contribute to the platform.
The setup below describes the **minimal local development workflow** and may evolve as the project grows.

To run the platform locally:

```bash
git clone https://github.com/libpulse/platform.git
cd platform
```
