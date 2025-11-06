"""
Semantic embedding module using sentence-transformers.
Converts text queries into dense vector representations for similarity search.
"""

from sentence_transformers import SentenceTransformer
import numpy as np
from typing import List, Union


class EmbeddingModel:
    """Wrapper for sentence-transformers model for text embeddings."""

    def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
        """
        Initialize the embedding model.

        Args:
            model_name: Name of the sentence-transformers model to use.
                       Default is 'all-MiniLM-L6-v2' (384 dimensions, fast, good quality)
        """
        print(f"Loading embedding model: {model_name}...")
        self.model = SentenceTransformer(model_name)
        self.dimension = self.model.get_sentence_embedding_dimension()
        print(f"Model loaded. Embedding dimension: {self.dimension}")

    def encode(self, texts: Union[str, List[str]], normalize: bool = True) -> np.ndarray:
        """
        Encode text(s) into embedding vector(s).

        Args:
            texts: Single text string or list of text strings
            normalize: Whether to normalize embeddings to unit length (for cosine similarity)

        Returns:
            numpy array of shape (n, dimension) where n is number of texts
        """
        if isinstance(texts, str):
            texts = [texts]

        embeddings = self.model.encode(
            texts,
            convert_to_numpy=True,
            normalize_embeddings=normalize,
            show_progress_bar=False
        )

        return embeddings

    def encode_batch(self, texts: List[str], batch_size: int = 32, normalize: bool = True) -> np.ndarray:
        """
        Encode a large batch of texts efficiently.

        Args:
            texts: List of text strings
            batch_size: Number of texts to process at once
            normalize: Whether to normalize embeddings

        Returns:
            numpy array of embeddings
        """
        return self.model.encode(
            texts,
            batch_size=batch_size,
            convert_to_numpy=True,
            normalize_embeddings=normalize,
            show_progress_bar=len(texts) > 100
        )

    def similarity(self, text1: str, text2: str) -> float:
        """
        Compute cosine similarity between two texts.

        Args:
            text1: First text
            text2: Second text

        Returns:
            Similarity score between 0 and 1 (1 = identical)
        """
        embeddings = self.encode([text1, text2], normalize=True)
        similarity = np.dot(embeddings[0], embeddings[1])
        return float(similarity)


# Global model instance (singleton pattern)
_model_instance = None


def get_embedding_model(model_name: str = "all-MiniLM-L6-v2") -> EmbeddingModel:
    """
    Get or create the global embedding model instance.

    Args:
        model_name: Name of the model to load

    Returns:
        EmbeddingModel instance
    """
    global _model_instance
    if _model_instance is None:
        _model_instance = EmbeddingModel(model_name)
    return _model_instance
