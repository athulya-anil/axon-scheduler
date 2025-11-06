"""
FAISS-based vector similarity search for semantic caching.
Stores job result embeddings and performs efficient nearest neighbor search.
"""

import faiss
import numpy as np
from typing import List, Tuple, Optional, Dict
from datetime import datetime, timedelta
import threading


class CacheEntry:
    """Represents a cached job result with metadata."""

    def __init__(self, query: str, result: str, embedding: np.ndarray, ttl_seconds: int = 3600):
        self.query = query
        self.result = result
        self.embedding = embedding
        self.created_at = datetime.now()
        self.ttl_seconds = ttl_seconds
        self.hit_count = 0

    def is_expired(self) -> bool:
        """Check if cache entry has exceeded its TTL."""
        return datetime.now() > self.created_at + timedelta(seconds=self.ttl_seconds)

    def to_dict(self) -> Dict:
        """Convert cache entry to dictionary."""
        return {
            "query": self.query,
            "result": self.result,
            "created_at": self.created_at.isoformat(),
            "ttl_seconds": self.ttl_seconds,
            "hit_count": self.hit_count,
            "age_seconds": (datetime.now() - self.created_at).total_seconds()
        }


class SemanticCache:
    """FAISS-based semantic cache for job results."""

    def __init__(self, dimension: int, similarity_threshold: float = 0.85):
        """
        Initialize the semantic cache.

        Args:
            dimension: Dimension of embedding vectors
            similarity_threshold: Minimum similarity score for cache hit (0-1)
        """
        self.dimension = dimension
        self.similarity_threshold = similarity_threshold

        # Use IndexFlatIP for inner product (cosine similarity with normalized vectors)
        self.index = faiss.IndexFlatIP(dimension)

        # Store cache entries (indexed by position in FAISS)
        self.entries: List[CacheEntry] = []

        # Thread lock for concurrent access
        self.lock = threading.Lock()

        # Statistics
        self.total_queries = 0
        self.cache_hits = 0
        self.cache_misses = 0

        print(f"Initialized semantic cache (dim={dimension}, threshold={similarity_threshold})")

    def add(self, query: str, result: str, embedding: np.ndarray, ttl_seconds: int = 3600) -> int:
        """
        Add a new entry to the cache.

        Args:
            query: Original query text
            result: Job result/output
            embedding: Query embedding vector
            ttl_seconds: Time-to-live in seconds

        Returns:
            Index of the added entry
        """
        with self.lock:
            # Create cache entry
            entry = CacheEntry(query, result, embedding, ttl_seconds)

            # Normalize embedding for cosine similarity
            normalized_embedding = embedding / np.linalg.norm(embedding)
            normalized_embedding = normalized_embedding.reshape(1, -1).astype('float32')

            # Add to FAISS index
            self.index.add(normalized_embedding)

            # Store entry
            self.entries.append(entry)

            return len(self.entries) - 1

    def search(self, query_embedding: np.ndarray, k: int = 1) -> Optional[Tuple[CacheEntry, float]]:
        """
        Search for similar cached results.

        Args:
            query_embedding: Embedding of the query
            k: Number of nearest neighbors to retrieve

        Returns:
            Tuple of (CacheEntry, similarity_score) if found, None otherwise
        """
        with self.lock:
            self.total_queries += 1

            if self.index.ntotal == 0:
                self.cache_misses += 1
                return None

            # Normalize query embedding
            normalized_embedding = query_embedding / np.linalg.norm(query_embedding)
            normalized_embedding = normalized_embedding.reshape(1, -1).astype('float32')

            # Search FAISS index
            distances, indices = self.index.search(normalized_embedding, min(k, self.index.ntotal))

            # Get best match
            best_idx = indices[0][0]
            similarity = float(distances[0][0])

            # Check if entry exists and is valid
            if best_idx >= len(self.entries):
                self.cache_misses += 1
                return None

            entry = self.entries[best_idx]

            # Check if expired
            if entry.is_expired():
                self.cache_misses += 1
                return None

            # Check similarity threshold
            if similarity < self.similarity_threshold:
                self.cache_misses += 1
                return None

            # Cache hit!
            self.cache_hits += 1
            entry.hit_count += 1

            return (entry, similarity)

    def cleanup_expired(self) -> int:
        """
        Remove expired entries from cache.

        Returns:
            Number of entries removed
        """
        with self.lock:
            # Find non-expired entries
            valid_entries = []
            valid_embeddings = []

            for entry in self.entries:
                if not entry.is_expired():
                    valid_entries.append(entry)
                    valid_embeddings.append(entry.embedding)

            removed_count = len(self.entries) - len(valid_entries)

            if removed_count > 0:
                # Rebuild index with valid entries only
                self.index.reset()
                self.entries = valid_entries

                if len(valid_embeddings) > 0:
                    embeddings_array = np.vstack(valid_embeddings).astype('float32')
                    # Normalize
                    norms = np.linalg.norm(embeddings_array, axis=1, keepdims=True)
                    embeddings_array = embeddings_array / norms
                    self.index.add(embeddings_array)

                print(f"Cleaned up {removed_count} expired entries")

            return removed_count

    def get_stats(self) -> Dict:
        """Get cache statistics."""
        with self.lock:
            hit_rate = self.cache_hits / self.total_queries if self.total_queries > 0 else 0.0

            return {
                "total_entries": len(self.entries),
                "total_queries": self.total_queries,
                "cache_hits": self.cache_hits,
                "cache_misses": self.cache_misses,
                "hit_rate": hit_rate,
                "hit_rate_percent": hit_rate * 100
            }

    def get_all_entries(self) -> List[Dict]:
        """Get all cache entries as dictionaries."""
        with self.lock:
            return [entry.to_dict() for entry in self.entries if not entry.is_expired()]

    def clear(self) -> int:
        """Clear all cache entries."""
        with self.lock:
            count = len(self.entries)
            self.index.reset()
            self.entries.clear()
            return count

    def size(self) -> int:
        """Get number of entries in cache."""
        with self.lock:
            return len(self.entries)


# Global cache instance
_cache_instance: Optional[SemanticCache] = None


def get_cache(dimension: int = 384, similarity_threshold: float = 0.85) -> SemanticCache:
    """
    Get or create the global semantic cache instance.

    Args:
        dimension: Embedding dimension
        similarity_threshold: Similarity threshold for cache hits

    Returns:
        SemanticCache instance
    """
    global _cache_instance
    if _cache_instance is None:
        _cache_instance = SemanticCache(dimension, similarity_threshold)
    return _cache_instance
