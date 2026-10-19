# Open-source embeddings without an OpenAI account

Select **OpenAI-compatible (Open-source models)** under the project's Embedding Provider. This connects directly to an embedding API; no LiteLLM proxy or OpenAI subscription is required. Existing LiteLLM configurations continue to work unchanged.

## Hugging Face Text Embeddings Inference (TEI)

TEI is an open-source server exposing `/v1/embeddings`. Run it on your own machine/server, or use a hosted TEI endpoint. Self-hosting does not require a payment card; hosted services may charge separately and require their own billing.

For a CPU Docker installation (initial startup downloads model weights):

```sh
docker run -p 127.0.0.1:8080:80 -v tei-data:/data ghcr.io/huggingface/text-embeddings-inference:cpu-1.9 --model-id BAAI/bge-m3
```

Configure:

| Field | Value |
|---|---|
| Provider | OpenAI-compatible (Open-source models) |
| API base URL | `http://localhost:8080/v1` when DecisionBox runs on the host |
| Authentication | No authentication for the local server |
| Model | `BAAI/bge-m3` |
| Dimensions | `1024` |

For DecisionBox in Docker, `localhost` means the DecisionBox container. Put TEI on the same Docker network and use `http://tei:80/v1` (service name `tei`), or configure an accessible host gateway. Both the API and agent must reach this address. A loopback-bound port is not a remote-server endpoint.

Use an endpoint provider's Bearer API token for authenticated hosted TEI. Enter its exact API prefix, including `/v1` if required; do not enter `/embeddings`. A Hugging Face feature-extraction URL returning a raw array is a different protocol and is not supported by this provider.

Models in the picker are suggestions, not downloads. Deploy the selected model on the server first. You can enter any other supported embedding model ID manually. If `/models` is unavailable, enter the deployed model ID and test an actual embedding. Qwen3-Embedding-0.6B is another option supported by TEI; use a suitable image/hardware from its documentation and review model-specific prompting requirements.

Smoke test from the same network as DecisionBox:

```sh
curl http://localhost:8080/v1/embeddings -H 'Content-Type: application/json' -d '{"model":"BAAI/bge-m3","input":["ข้อมูลทรัพย์สินทางปัญญา"],"encoding_format":"float"}'
```

Save the provider settings and rebuild/reindex the schema and discovery embeddings when changing model or dimensions. Old vectors cannot be meaningfully searched with a different model, even if vector lengths match. These changes require rebuilding the API and agent images; pulling an unchanged upstream image will not include them.

## Ollama troubleshooting

Ollama already supports embedding through `/api/embed`. It needs an embedding-capable model, not just a chat model:

```sh
ollama pull bge-m3
curl http://localhost:11434/api/embed -H 'Content-Type: application/json' -d '{"model":"bge-m3","input":["ทดสอบ embedding"]}'
```

Set the Ollama host to a URL reachable by both DecisionBox API and agent. A 404 model-not-found needs a model pull on that server; connection refused usually needs host/network correction. Keep the full error message when diagnosing timeouts or unsupported models. This guide does not establish the cause of any particular user's failure.

References:
- https://huggingface.co/docs/text-embeddings-inference/quick_tour
- https://huggingface.co/docs/text-embeddings-inference/supported_models
- https://huggingface.co/BAAI/bge-m3
