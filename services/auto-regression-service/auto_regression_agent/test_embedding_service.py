#!/usr/bin/env python3
"""
Test script for local embedding service
Tests both the embedding service and compares with OpenAI format
"""

import requests
import json
import sys

def test_local_embedding_service(base_url="http://localhost:8000"):
    """Test the local embedding service"""
    print("=" * 60)
    print("Testing Local Embedding Service")
    print("=" * 60)
    
    # Test 1: Health check
    print("\n1. Testing health endpoint...")
    try:
        response = requests.get(f"{base_url}/health")
        print(f"   Status: {response.status_code}")
        print(f"   Response: {response.json()}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False
    
    # Test 2: Root endpoint (model info)
    print("\n2. Testing root endpoint (model info)...")
    try:
        response = requests.get(f"{base_url}/")
        data = response.json()
        print(f"   Status: {response.status_code}")
        print(f"   Model: {data.get('model')}")
        print(f"   Embedding Dimension: {data.get('embedding_dimension')}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False
    
    # Test 3: Single text embedding
    print("\n3. Testing single text embedding...")
    try:
        payload = {
            "input": "Hello, this is a test sentence for embedding generation.",
            "model": "all-MiniLM-L6-v2"
        }
        response = requests.post(f"{base_url}/v1/embeddings", json=payload)
        data = response.json()
        print(f"   Status: {response.status_code}")
        print(f"   Model: {data.get('model')}")
        print(f"   Number of embeddings: {len(data.get('data', []))}")
        if data.get('data'):
            embedding = data['data'][0]['embedding']
            print(f"   Embedding dimension: {len(embedding)}")
            print(f"   First 5 values: {embedding[:5]}")
            print(f"   Usage: {data.get('usage')}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False
    
    # Test 4: Batch embeddings
    print("\n4. Testing batch embeddings...")
    try:
        payload = {
            "input": [
                "First test sentence",
                "Second test sentence",
                "Third test sentence"
            ],
            "model": "all-MiniLM-L6-v2"
        }
        response = requests.post(f"{base_url}/v1/embeddings", json=payload)
        data = response.json()
        print(f"   Status: {response.status_code}")
        print(f"   Number of embeddings: {len(data.get('data', []))}")
        for i, item in enumerate(data.get('data', [])):
            print(f"   Embedding {i}: dimension={len(item['embedding'])}, index={item['index']}")
        print(f"   Usage: {data.get('usage')}")
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False
    
    # Test 5: OpenAI compatibility check
    print("\n5. Testing OpenAI API compatibility...")
    try:
        payload = {
            "input": "Testing OpenAI compatibility",
            "model": "all-MiniLM-L6-v2",
            "encoding_format": "float"
        }
        response = requests.post(f"{base_url}/v1/embeddings", json=payload)
        data = response.json()
        
        # Check OpenAI-compatible response format
        assert data.get('object') == 'list', "Missing 'object' field"
        assert 'data' in data, "Missing 'data' field"
        assert 'model' in data, "Missing 'model' field"
        assert 'usage' in data, "Missing 'usage' field"
        assert data['data'][0].get('object') == 'embedding', "Invalid data object type"
        assert 'embedding' in data['data'][0], "Missing embedding field"
        assert 'index' in data['data'][0], "Missing index field"
        
        print(f"   ✅ Response format is OpenAI-compatible")
        print(f"   Response structure:")
        print(f"      - object: {data.get('object')}")
        print(f"      - data[0].object: {data['data'][0].get('object')}")
        print(f"      - data[0].index: {data['data'][0].get('index')}")
        print(f"      - model: {data.get('model')}")
        print(f"      - usage: {data.get('usage')}")
    except AssertionError as e:
        print(f"   ❌ Compatibility Error: {e}")
        return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False
    
    print("\n" + "=" * 60)
    print("✅ All tests passed!")
    print("=" * 60)
    return True

if __name__ == "__main__":
    base_url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8000"
    success = test_local_embedding_service(base_url)
    sys.exit(0 if success else 1)

