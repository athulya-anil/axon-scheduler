"""
Axon Semantic Cache Service
FastAPI-based REST API for semantic caching using sentence embeddings and FAISS.
"""

from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel, Field
from typing import Optional, List, Dict
import uvicorn
from contextlib import asynccontextmanager
import asyncio

from embeddings import get_embedding_model
from faiss_index import get_cache


# Pydantic models for API
class CacheRequest(BaseModel):
    """Request to add an entry to the cache."""
    query: str = Field(..., description="Query text to cache", min_length=1)
    result: str = Field(..., description="Result to cache", min_length=1)
    ttl_seconds: int = Field(default=3600, description="Time-to-live in seconds", ge=1)


class SearchRequest(BaseModel):
    """Request to search the cache."""
    query: str = Field(..., description="Query text to search", min_length=1)


class SearchResponse(BaseModel):
    """Response from cache search."""
    hit: bool = Field(description="Whether a cache hit occurred")
    query: Optional[str] = Field(None, description="Original cached query")
    result: Optional[str] = Field(None, description="Cached result")
    similarity: Optional[float] = Field(None, description="Similarity score (0-1)")
    metadata: Optional[Dict] = Field(None, description="Cache entry metadata")


class StatsResponse(BaseModel):
    """Cache statistics."""
    total_entries: int
    total_queries: int
    cache_hits: int
    cache_misses: int
    hit_rate: float
    hit_rate_percent: float


# Background task for cache cleanup
async def cleanup_task():
    """Periodically clean up expired cache entries."""
    cache = get_cache()
    while True:
        await asyncio.sleep(300)  # Run every 5 minutes
        try:
            removed = cache.cleanup_expired()
            if removed > 0:
                print(f"Cleanup: removed {removed} expired entries")
        except Exception as e:
            print(f"Cleanup error: {e}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifecycle manager for the FastAPI application."""
    # Startup
    print("Starting Axon Semantic Cache Service...")
    print("Loading embedding model...")
    get_embedding_model()  # Preload model
    print("Initializing cache...")
    get_cache()
    print("Starting background cleanup task...")
    cleanup = asyncio.create_task(cleanup_task())

    yield

    # Shutdown
    print("Shutting down cache service...")
    cleanup.cancel()


# Initialize FastAPI app
app = FastAPI(
    title="Axon Semantic Cache",
    description="Semantic caching service using sentence embeddings and FAISS for similarity search",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/")
def root():
    """Health check endpoint."""
    return {
        "service": "Axon Semantic Cache",
        "status": "ok",
        "version": "1.0.0"
    }


@app.get("/health")
def health():
    """Detailed health check."""
    cache = get_cache()
    model = get_embedding_model()

    return {
        "status": "healthy",
        "cache_size": cache.size(),
        "embedding_dimension": model.dimension,
        "model": "all-MiniLM-L6-v2"
    }


@app.post("/cache", status_code=201)
def add_to_cache(request: CacheRequest):
    """
    Add a new entry to the semantic cache.

    Args:
        request: Cache request with query, result, and TTL

    Returns:
        Cache entry ID and metadata
    """
    try:
        # Get model and cache
        model = get_embedding_model()
        cache = get_cache()

        # Generate embedding
        embedding = model.encode(request.query, normalize=True)

        # Add to cache
        entry_id = cache.add(
            query=request.query,
            result=request.result,
            embedding=embedding,
            ttl_seconds=request.ttl_seconds
        )

        return {
            "entry_id": entry_id,
            "query": request.query,
            "ttl_seconds": request.ttl_seconds,
            "message": "Entry added to cache"
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to add to cache: {str(e)}")


@app.post("/search", response_model=SearchResponse)
def search_cache(request: SearchRequest):
    """
    Search the semantic cache for similar queries.

    Args:
        request: Search request with query text

    Returns:
        Cache hit information or miss
    """
    try:
        # Get model and cache
        model = get_embedding_model()
        cache = get_cache()

        # Generate query embedding
        query_embedding = model.encode(request.query, normalize=True)

        # Search cache
        result = cache.search(query_embedding)

        if result is None:
            return SearchResponse(hit=False)

        entry, similarity = result

        return SearchResponse(
            hit=True,
            query=entry.query,
            result=entry.result,
            similarity=similarity,
            metadata=entry.to_dict()
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")


@app.get("/stats", response_model=StatsResponse)
def get_stats():
    """Get cache statistics including hit rate."""
    try:
        cache = get_cache()
        stats = cache.get_stats()
        return StatsResponse(**stats)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get stats: {str(e)}")


@app.get("/entries")
def get_entries():
    """Get all cache entries (non-expired)."""
    try:
        cache = get_cache()
        entries = cache.get_all_entries()

        return {
            "count": len(entries),
            "entries": entries
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get entries: {str(e)}")


@app.delete("/cache")
def clear_cache():
    """Clear all cache entries."""
    try:
        cache = get_cache()
        count = cache.clear()

        return {
            "message": f"Cache cleared",
            "entries_removed": count
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to clear cache: {str(e)}")


@app.post("/cleanup")
def manual_cleanup():
    """Manually trigger cleanup of expired entries."""
    try:
        cache = get_cache()
        removed = cache.cleanup_expired()

        return {
            "message": "Cleanup completed",
            "entries_removed": removed
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Cleanup failed: {str(e)}")


@app.post("/similarity")
def compute_similarity(text1: str, text2: str):
    """
    Compute semantic similarity between two texts.

    Args:
        text1: First text
        text2: Second text

    Returns:
        Similarity score (0-1)
    """
    try:
        model = get_embedding_model()
        similarity = model.similarity(text1, text2)

        return {
            "text1": text1,
            "text2": text2,
            "similarity": similarity
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Similarity computation failed: {str(e)}")


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info"
    )
