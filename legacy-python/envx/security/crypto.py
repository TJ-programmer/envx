"""Encryption helpers."""

from __future__ import annotations

from cryptography.fernet import Fernet, InvalidToken

from envx.core.exceptions import SecretKeyError
from envx.security.keyring import LocalKeyProvider

ENCRYPTED_PREFIX = "enc:"


class CryptoManager:
    """Encrypt and decrypt values with a project-local key."""

    def __init__(self, key_provider: LocalKeyProvider) -> None:
        self.key_provider = key_provider

    def ensure_key(self) -> None:
        """Ensure a local key exists."""
        self.key_provider.ensure_key()

    def encrypt(self, value: str) -> str:
        """Encrypt a plaintext secret value."""
        fernet = Fernet(self.key_provider.load_key())
        return ENCRYPTED_PREFIX + fernet.encrypt(value.encode("utf-8")).decode("utf-8")

    def decrypt(self, value: str) -> str:
        """Decrypt an encrypted value or return plaintext values unchanged."""
        if not value.startswith(ENCRYPTED_PREFIX):
            return value

        fernet = Fernet(self.key_provider.load_key())
        try:
            return fernet.decrypt(value[len(ENCRYPTED_PREFIX) :].encode("utf-8")).decode(
                "utf-8"
            )
        except InvalidToken as exc:
            raise SecretKeyError(
                "Failed to decrypt a stored secret. The key may be missing or mismatched."
            ) from exc
