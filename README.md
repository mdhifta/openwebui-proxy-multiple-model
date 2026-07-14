# Configuration

```env
PORT=9090
```

> **Note**
> The proxy only supports port 9090. If you want to change the port, make sure to update port in .env.

```env
VERTEXAI_PROJECT={vertex-project-id}
```
```
VERTEXAI_LOCATION={location}
```
Example:

- `global`
- `us-central1`

```
VERTEXAI_ANTHROPIC_VERSION={vertexai-anthropic-version}
```
Example:

- `vertex-2023-10-16`


```env
VERTEXAI_AVAILABLE_MODELS={model-list-vertex}
```

Example:

```json
[
  "google/gemini-2.5-pro",
  "google/gemini-2.5-flash",
  "google/gemini-2.5-flash-image",
  "gemini-3.1-pro-preview",
  "gemini-3-flash-preview",
  "claude-sonnet-4-6"
]
```

> **Note**
> These models must be defined manually because Vertex AI does not provide an API to retrieve the list of models available in your project.

```env
OPENAI_API_KEY={openai-api-key}
```

Set this if you want to enable OpenAI support.

## Google Cloud Credentials

Save your Google Cloud service account credentials in the project root directory (`./`) using the filename:

```
credential-gcp.json
```

## Run

Build the image:

```bash
docker compose build
```

Start the service:

```bash
docker compose up
```

The OpenAI-compatible endpoint will be available at:

```
http://localhost:9090/v1
```
---
# Architecture Flow Schema
![Proxy Openwebui Acrhitecture](https://github.com/user-attachments/assets/216c1901-2439-4575-8b2a-6733e9cc1365)


---

# Open WebUI Setup

## 1. Admin Panel → Connections

![Open WebUI Connections](https://github.com/user-attachments/assets/9911a768-3e1a-421f-b716-a9b3a6ce98db)

## 2. Admin Panel → Models

![Open WebUI Models](https://github.com/user-attachments/assets/e6f1affc-9ff5-41c9-b4d6-f901d8b89d42)

# Documentation 
https://medium.com/@mdhiftaa/openwebui-proxy-multiple-models-b072e734385d?sharedUserId=mdhiftaa
