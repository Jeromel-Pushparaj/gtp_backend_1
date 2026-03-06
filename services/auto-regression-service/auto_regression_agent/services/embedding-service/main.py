#!/usr/bin/env python3
"""
Local Embedding Service using Sentence Transformers
Provides OpenAI-compatible API for generating embeddings locally
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from typing import List, Union
import uvicorn
import logging
import os

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Local Embedding Service",
    description="OpenAI-compatible embedding API using Sentence Transformers",
    version="1.0.0"
)

# Load model on startup
MODEL_NAME = os.getenv("EMBEDDING_MODEL", "all-MiniLM-L6-v2")
logger.info(f"Loading model: {MODEL_NAME}")
model = SentenceTransformer(MODEL_NAME)
logger.info(f"Model loaded successfully. Embedding dimension: {model.get_sentence_embedding_dimension()}")


class EmbeddingRequest(BaseModel):
    """OpenAI-compatible embedding request"""
    input: Union[str, List[str]]
    model: str = MODEL_NAME
    encoding_format: str = "float"


class EmbeddingData(BaseModel):
    """Single embedding result"""
    object: str = "embedding"
    embedding: List[float]
    index: int


class EmbeddingResponse(BaseModel):
    """OpenAI-compatible embedding response"""
    object: str = "list"
    data: List[EmbeddingData]
    model: str
    usage: dict


@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "model": MODEL_NAME,
        "embedding_dimension": model.get_sentence_embedding_dimension()
    }


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy"}


@app.post("/v1/embeddings")
@app.post("/embeddings")
async def create_embeddings(request: EmbeddingRequest):
    """
    Generate embeddings for input text(s)
    OpenAI-compatible endpoint
    """
    try:
        # Handle both single string and list of strings
        if isinstance(request.input, str):
            texts = [request.input]
        else:
            texts = request.input
        
        if not texts:
            raise HTTPException(status_code=400, detail="Input cannot be empty")
        
        # Generate embeddings
        logger.info(f"Generating embeddings for {len(texts)} text(s)")
        embeddings = model.encode(texts, convert_to_numpy=True)
        
        # Convert to list format
        if len(embeddings.shape) == 1:
            embeddings = [embeddings]
        
        # Build response in OpenAI format
        data = []
        for idx, embedding in enumerate(embeddings):
            data.append(EmbeddingData(
                embedding=embedding.tolist(),
                index=idx
            ))
        
        response = EmbeddingResponse(
            data=data,
            model=request.model,
            usage={
                "prompt_tokens": sum(len(text.split()) for text in texts),
                "total_tokens": sum(len(text.split()) for text in texts)
            }
        )
        
        return response
    
    except Exception as e:
        logger.error(f"Error generating embeddings: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/models")
async def list_models():
    """List available models (OpenAI-compatible)"""
    return {
        "object": "list",
        "data": [
            {
                "id": MODEL_NAME,
                "object": "model",
                "owned_by": "sentence-transformers",
                "permission": []
            }
        ]
    }


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    
    logger.info(f"Starting embedding service on {host}:{port}")
    uvicorn.run(app, host=host, port=port)

