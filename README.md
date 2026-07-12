This proxy was created by M Dhifta to bridge OpenWebUI with Claude, Gemini, and OpenAI. Below are the requirements needed to ensure the proxy runs properly:

```env
PORT=9090
# The proxy only supports port **9090**. If you want to change the port, make sure to update the `EXPOSE 9090` instruction in the Dockerfile as well.

VERTEXAI_PROJECT={vertex-project-id}
VERTEXAI_LOCATION={location}
# Example: global, us-central1, etc.

VERTEXAI_AVAILABLE_MODELS={model-list-vertex}
# Example:
# ["google/gemini-2.5-pro","google/gemini-2.5-flash","google/gemini-2.5-flash-image","gemini-3.1-pro-preview","gemini-3-flash-preview","claude-sonnet-4-6"]

# These models must be defined manually because Vertex AI does not provide an API or function
# to retrieve the list of models available in your project.

OPENAI_API_KEY={openai-api-key}
# Set this if you want to use or enable OpenAI.

Make sure you have saved your Google Cloud service account credentials in the project root directory (`./`) with the filename `credential-gcp.json`.

# How to run
-- docker compose build
-- docker compose up

http://localhost:9090/v1

# Openwebui Setup
Admin Panel -> connetions
<img width="1907" height="964" alt="Screenshot from 2026-07-12 11-29-56" src="https://github.com/user-attachments/assets/9911a768-3e1a-421f-b716-a9b3a6ce98db" />

Admin Panel -> Models
<img width="1907" height="964" alt="Screenshot from 2026-07-12 11-47-23" src="https://github.com/user-attachments/assets/e6f1affc-9ff5-41c9-b4d6-f901d8b89d42" />

```
