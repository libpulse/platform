Services

The services/ directory contains all backend services that power the LibPulse platform.
Each service is isolated, versioned, and can be deployed independently in the future.

Currently available:

📡 api/

The primary HTTP API service.
Responsible for:
	•	Event ingestion (logs, errors, metrics)
	•	Project & API key management
	•	User consent handling
	•	Dashboard-friendly read APIs
	•	OpenAPI specification (api/openapi.yaml)

Implemented in Go (Gin).