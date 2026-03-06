# Local Embedding Service

OpenAI-compatible embedding API using Sentence Transformers for local, privacy-preserving embeddings.

## Features

- 🔒 **Privacy-First**: All embeddings generated locally, no data sent to external APIs
- 🚀 **Fast**: Uses optimized sentence-transformers models
- 🔌 **OpenAI-Compatible**: Drop-in replacement for OpenAI embeddings API
- 🐳 **Docker Ready**: Easy deployment with Docker

## Supported Models

- `all-MiniLM-L6-v2` (default) - 384 dimensions, fast and efficient
- `all-mpnet-base-v2` - 768 dimensions, higher quality
- `paraphrase-multilingual-MiniLM-L12-v2` - 384 dimensions, multilingual

## Quick Start

### Using Docker

```bash
docker build -t embedding-service .
docker run -p 8000:8000 embedding-service
```

### Using Python

```bash
pip install -r requirements.txt
python main.py
```

## API Usage

### Generate Embeddings

```bash
curl -X POST http://localhost:8000/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello, world!",
    "model": "all-MiniLM-L6-v2"
  }'
```

### Batch Embeddings

```bash
curl -X POST http://localhost:8000/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": ["First text", "Second text", "Third text"],
    "model": "all-MiniLM-L6-v2"
  }'
```

### Response Format

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.123, -0.456, ...],
      "index": 0
    }
  ],
  "model": "all-MiniLM-L6-v2",
  "usage": {
    "prompt_tokens": 3,
    "total_tokens": 3
  }
}
```

## Configuration

Environment variables:

- `EMBEDDING_MODEL`: Model name (default: `all-MiniLM-L6-v2`)
- `PORT`: Server port (default: `8000`)
- `HOST`: Server host (default: `0.0.0.0`)

## Model Dimensions

| Model | Dimensions | Use Case |
|-------|-----------|----------|
| all-MiniLM-L6-v2 | 384 | Fast, general purpose |
| all-mpnet-base-v2 | 768 | Higher quality |
| paraphrase-multilingual-MiniLM-L12-v2 | 384 | Multilingual |

## Integration with OpenTest

The service is automatically integrated with the OpenTest platform. Update your `.env`:

```bash
EMBEDDING_PROVIDER=local
EMBEDDING_MODEL=all-MiniLM-L6-v2
```

## Performance

- **Throughput**: ~1000 embeddings/second on CPU
- **Latency**: ~10ms per embedding (single)
- **Memory**: ~500MB for all-MiniLM-L6-v2

## License

Apache 2.0

