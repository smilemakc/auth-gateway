"""Auth Gateway user importer via REST API bulk import endpoint."""

from __future__ import annotations

import asyncio
import logging
from typing import Any

import httpx

logger = logging.getLogger(__name__)

from . import (
    ConflictStrategy,
    ImportResult,
    ImportStats,
    ImportStatus,
    PasswordConfig,
    PasswordStrategy,
    RolesConfig,
    SourceUser,
    TargetConfig,
)


class AuthGatewayImporter:
    def __init__(
        self,
        config: TargetConfig,
        password_config: PasswordConfig,
        conflict_strategy: ConflictStrategy,
        workers: int = 4,
        roles_config: RolesConfig | None = None,
    ):
        self.config = config
        self.password_config = password_config
        self.conflict_strategy = conflict_strategy
        self.workers = workers
        self.roles_config = roles_config
        self.stats = ImportStats()
        self._client: httpx.AsyncClient | None = None

    async def __aenter__(self):
        self._client = httpx.AsyncClient(
            base_url=self.config.base_url,
            headers={
                "X-API-Key": self.config.api_key,
                "X-Application-ID": self.config.application_id,
                "Content-Type": "application/json",
            },
            timeout=30.0,
        )
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._client:
            await self._client.aclose()

    async def import_users(self, users: list[SourceUser]) -> list[ImportResult]:
        if not self._client:
            raise RuntimeError("Importer must be used as async context manager")

        results: list[ImportResult] = []
        batch_size = 50
        batches = [users[i : i + batch_size] for i in range(0, len(users), batch_size)]

        semaphore = asyncio.Semaphore(self.workers)

        async def process_batch(batch: list[SourceUser]) -> list[ImportResult]:
            async with semaphore:
                return await self.import_batch(batch)

        tasks = [process_batch(batch) for batch in batches]
        batch_results = await asyncio.gather(*tasks, return_exceptions=True)

        for batch_result in batch_results:
            if isinstance(batch_result, Exception):
                self.stats.errors += 1
                continue
            results.extend(batch_result)

        return results

    def _build_import_entry(self, user: SourceUser) -> dict[str, Any]:
        entry: dict[str, Any] = {
            "is_active": user.is_active,
            "email_verified": user.email_verified,
            "phone_verified": user.phone_verified,
        }

        if user.email:
            entry["email"] = user.email

        if user.phone:
            entry["phone"] = user.phone

        # Username: explicit → email prefix → phone digits
        username = user.username
        if not username and user.email:
            username = user.email.split("@")[0]
        if not username and user.phone:
            username = user.phone.lstrip("+")
        if username:
            entry["username"] = username

        if user.full_name:
            entry["full_name"] = user.full_name

        if self.password_config.strategy == PasswordStrategy.TRANSFER and user.password_hash:
            entry["password_hash_import"] = user.password_hash

        if user.id:
            entry["id"] = user.id

        # Map source role IDs to Auth Gateway role names
        if self.roles_config and self.roles_config.enabled and user.roles:
            app_roles = [
                self.roles_config.mapping[r]
                for r in user.roles
                if r in self.roles_config.mapping
            ]
            if app_roles:
                entry["app_roles"] = app_roles

        return entry

    async def import_batch(self, batch: list[SourceUser]) -> list[ImportResult]:
        if not self._client:
            raise RuntimeError("HTTP client not initialized")

        entries = [self._build_import_entry(user) for user in batch]
        payload = {
            "users": entries,
            "on_conflict": self.conflict_strategy.value,
        }

        try:
            response = await self._client.post("/api/admin/users/import", json=payload)
            response.raise_for_status()

            data = response.json()
            imported = data.get("imported", 0)
            skipped = data.get("skipped", 0)
            updated = data.get("updated", 0)
            errors = data.get("errors", 0)
            details = data.get("details", [])

            self.stats.total += len(batch)
            self.stats.created += imported
            self.stats.skipped += skipped
            self.stats.updated += updated
            self.stats.errors += errors

            results: list[ImportResult] = []
            for i, user in enumerate(batch):
                detail = details[i] if i < len(details) else {}
                status_str = detail.get("status", "created")
                ag_id = detail.get("user_id")
                note = detail.get("reason", "")

                if status_str in ("created", "imported"):
                    status = ImportStatus.CREATED
                elif status_str == "existing":
                    status = ImportStatus.EXISTING
                elif status_str == "skipped":
                    status = ImportStatus.SKIPPED
                elif status_str == "updated":
                    status = ImportStatus.UPDATED
                elif status_str == "error":
                    status = ImportStatus.ERROR
                else:
                    status = ImportStatus.CREATED

                results.append(
                    ImportResult(source_id=user.id, ag_id=ag_id, status=status, note=note)
                )

            return results

        except httpx.HTTPStatusError as e:
            error_msg = f"HTTP {e.response.status_code}: {e.response.text[:500]}"
            logger.error("Batch import failed: %s", error_msg)
            for user in batch:
                self.stats.errors += 1
                self.stats.total += 1
            return [
                ImportResult(source_id=user.id, status=ImportStatus.ERROR, note=error_msg)
                for user in batch
            ]

        except Exception as e:
            error_msg = f"Request failed: {str(e)}"
            logger.error("Batch import exception: %s", error_msg)
            for user in batch:
                self.stats.errors += 1
                self.stats.total += 1
            return [
                ImportResult(source_id=user.id, status=ImportStatus.ERROR, note=error_msg)
                for user in batch
            ]
