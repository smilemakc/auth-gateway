"""Password hash detection and utility functions."""

from __future__ import annotations

import re


class DuplicateUserError(Exception):
    pass


def detect_algorithm(hash_value: str) -> str:
    if not hash_value:
        return "unknown"

    if isinstance(hash_value, bytes):
        hash_value = hash_value.decode("utf-8", errors="ignore")

    if hash_value.startswith("$2b$") or hash_value.startswith("$2a$") or hash_value.startswith("$2y$"):
        return "bcrypt"

    if hash_value.startswith("$argon2"):
        return "argon2"

    if hash_value.startswith("$scrypt"):
        return "scrypt"

    if hash_value.startswith("pbkdf2"):
        return "pbkdf2"

    hash_clean = re.sub(r"[^a-fA-F0-9]", "", hash_value)

    if len(hash_clean) == 64:
        return "sha256"

    if len(hash_clean) == 32:
        return "md5"

    if len(hash_clean) == 40:
        return "sha1"

    return "unknown"


def detect_algorithm_from_sample(hashes: list[str]) -> str:
    if not hashes:
        return "unknown"

    algorithms = []
    for hash_value in hashes:
        algorithm = detect_algorithm(hash_value)
        algorithms.append(algorithm)

    return max(set(algorithms), key=algorithms.count)


def is_bcrypt_compatible(hash_value: str) -> bool:
    if not hash_value:
        return False

    if isinstance(hash_value, bytes):
        hash_value = hash_value.decode("utf-8", errors="ignore")

    if hash_value.startswith("$2b$") or hash_value.startswith("$2a$") or hash_value.startswith("$2y$"):
        bcrypt_pattern = r"^\$2[aby]\$\d{2}\$[./A-Za-z0-9]{53}$"
        return bool(re.match(bcrypt_pattern, hash_value))

    return False
