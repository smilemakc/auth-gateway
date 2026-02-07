"""Migration verification - checks data integrity after import."""

from __future__ import annotations

import random
from typing import Any

import httpx

from . import SourceConfig, TargetConfig, VerificationReport


def _connect_source(config: SourceConfig):
    if config.type == "postgresql":
        import psycopg2
        return psycopg2.connect(
            host=config.host,
            port=config.port,
            dbname=config.database,
            user=config.user,
            password=config.password,
        )
    elif config.type == "mysql":
        import pymysql
        return pymysql.connect(
            host=config.host,
            port=config.port,
            database=config.database,
            user=config.user,
            password=config.password,
        )
    else:
        raise ValueError(f"Unsupported database type: {config.type}")


class MigrationVerifier:
    def __init__(self, source_config: SourceConfig, target_config: TargetConfig):
        self.source_config = source_config
        self.target_config = target_config
        self._http_client: httpx.AsyncClient | None = None
        self._conn = None

    async def __aenter__(self):
        self._http_client = httpx.AsyncClient(
            base_url=self.target_config.base_url,
            headers={
                "X-API-Key": self.target_config.api_key,
                "X-Application-ID": self.target_config.application_id,
            },
            timeout=30.0,
        )
        self._conn = _connect_source(self.source_config)
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._http_client:
            await self._http_client.aclose()
        if self._conn:
            self._conn.close()

    async def verify(self) -> VerificationReport:
        if not self._http_client or not self._conn:
            raise RuntimeError("Verifier must be used as async context manager")

        report = VerificationReport()

        source_count = self._count_source_users()
        ag_count = await self._count_ag_users()
        report.add_check("User count", source_count, ag_count)

        app_profiles_count = await self._count_app_profiles()
        report.add_check("App profiles count", source_count, app_profiles_count)

        sample_users = self._get_sample_users(min(100, source_count))
        for user in sample_users:
            user_id = str(user["id"])
            email = user["email"]

            ag_user = await self._find_in_ag(user_id)
            if not ag_user:
                report.add_missing(user_id, email)
            elif ag_user.get("email") != email:
                report.add_mismatch(
                    user_id, "email", email, ag_user.get("email", "")
                )

        return report

    def _count_source_users(self) -> int:
        cursor = self._conn.cursor()
        cursor.execute(f"SELECT COUNT(*) FROM {self.source_config.users_table}")
        result = cursor.fetchone()
        cursor.close()
        return result[0] if result else 0

    def _get_sample_users(self, sample_size: int) -> list[dict[str, Any]]:
        id_col = self.source_config.columns.get("id", "id")
        email_col = self.source_config.columns.get("email", "email")

        cursor = self._conn.cursor()
        if self.source_config.type == "postgresql":
            cursor.execute(
                f"SELECT {id_col} as id, {email_col} as email "
                f"FROM {self.source_config.users_table} "
                f"ORDER BY RANDOM() LIMIT %s",
                (sample_size,),
            )
        else:
            cursor.execute(
                f"SELECT {id_col} as id, {email_col} as email "
                f"FROM {self.source_config.users_table} "
                f"ORDER BY RAND() LIMIT %s",
                (sample_size,),
            )
        columns = [desc[0] for desc in cursor.description]
        rows = [dict(zip(columns, row)) for row in cursor.fetchall()]
        cursor.close()
        return rows

    async def _count_ag_users(self) -> int:
        try:
            response = await self._http_client.get("/api/admin/users", params={"per_page": 1})
            response.raise_for_status()
            return response.json().get("total", 0)
        except Exception:
            return 0

    async def _count_app_profiles(self) -> int:
        try:
            response = await self._http_client.get(
                f"/api/admin/applications/{self.target_config.application_id}/users",
                params={"per_page": 1},
            )
            response.raise_for_status()
            return response.json().get("total", 0)
        except Exception:
            return 0

    async def _find_in_ag(self, user_id: str) -> dict[str, Any] | None:
        try:
            response = await self._http_client.get(f"/api/admin/users/{user_id}")
            if response.status_code == 404:
                return None
            response.raise_for_status()
            return response.json()
        except Exception:
            return None
