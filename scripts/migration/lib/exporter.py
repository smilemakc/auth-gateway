"""User exporter from source PostgreSQL or MySQL database."""

from __future__ import annotations

import re
from typing import Iterator, Optional

try:
    import psycopg2
    import psycopg2.extras
except ImportError:
    psycopg2 = None

try:
    import pymysql
    import pymysql.cursors
except ImportError:
    pymysql = None

from . import SourceConfig, SourceUser


class UserExporter:
    def __init__(self, config: SourceConfig):
        self.config = config
        self.conn = None
        self._connect()

    def _connect(self):
        if self.config.type == "postgresql":
            if psycopg2 is None:
                raise ImportError("psycopg2 is required for PostgreSQL. Install with: pip install psycopg2-binary")

            self.conn = psycopg2.connect(
                host=self.config.host,
                port=self.config.port,
                database=self.config.database,
                user=self.config.user,
                password=self.config.password,
                sslmode="require" if self.config.ssl else "disable",
            )

        elif self.config.type == "mysql":
            if pymysql is None:
                raise ImportError("pymysql is required for MySQL. Install with: pip install pymysql")

            self.conn = pymysql.connect(
                host=self.config.host,
                port=self.config.port,
                database=self.config.database,
                user=self.config.user,
                password=self.config.password,
                ssl={"require": True} if self.config.ssl else None,
            )

        else:
            raise ValueError(f"Unsupported database type: {self.config.type}")

    def export_users(self) -> Iterator[SourceUser]:
        cols = self.config.columns

        select_parts = []
        for field, column_name in cols.items():
            select_parts.append(f"{column_name} AS {field}")

        select_clause = ", ".join(select_parts)
        order_by = cols.get("created_at", "created_at")

        query = f"SELECT {select_clause} FROM {self.config.users_table} ORDER BY {order_by} ASC"

        if self.config.type == "postgresql":
            cursor = self.conn.cursor(name="user_export_cursor", cursor_factory=psycopg2.extras.RealDictCursor)
            cursor.itersize = self.config.batch_size
            cursor.execute(query)

            for row in cursor:
                yield self._map_to_source_user(row)

            cursor.close()

        elif self.config.type == "mysql":
            cursor = self.conn.cursor(pymysql.cursors.SSCursor)
            cursor.execute(query)

            columns = [desc[0] for desc in cursor.description]

            for row_tuple in cursor:
                row = dict(zip(columns, row_tuple))
                yield self._map_to_source_user(row)

            cursor.close()

    def _map_to_source_user(self, row: dict) -> SourceUser:
        return SourceUser(
            id=str(row.get("id", "")),
            email=row.get("email") or "",
            username=row.get("username"),
            password_hash=row.get("password_hash"),
            full_name=row.get("full_name"),
            phone=row.get("phone"),
            is_active=bool(row.get("is_active", True)),
            created_at=row.get("created_at"),
        )

    def count_users(self) -> int:
        cursor = self.conn.cursor()
        cursor.execute(f"SELECT COUNT(*) FROM {self.config.users_table}")
        result = cursor.fetchone()
        cursor.close()
        return result[0] if result else 0

    def detect_id_type(self) -> str:
        id_column = self.config.columns.get("id", "id")

        if self.config.type == "postgresql":
            cursor = self.conn.cursor()
            cursor.execute(
                """
                SELECT data_type, udt_name
                FROM information_schema.columns
                WHERE table_name = %s AND column_name = %s
                """,
                (self.config.users_table, id_column)
            )
            row = cursor.fetchone()
            cursor.close()

            if row:
                data_type, udt_name = row
                if data_type == "uuid" or udt_name == "uuid":
                    return "uuid"
                elif data_type in ("integer", "bigint", "smallint"):
                    return "integer"

        elif self.config.type == "mysql":
            cursor = self.conn.cursor()
            cursor.execute(
                """
                SELECT DATA_TYPE, COLUMN_TYPE
                FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %s AND COLUMN_NAME = %s
                """,
                (self.config.database, self.config.users_table, id_column)
            )
            row = cursor.fetchone()
            cursor.close()

            if row:
                data_type, column_type = row
                if "char(36)" in column_type.lower() or "uuid" in column_type.lower():
                    return "uuid"
                elif data_type in ("int", "bigint", "smallint", "tinyint"):
                    return "integer"

        return "uuid"

    def detect_password_algorithm(self, sample_size: int = 10) -> str:
        password_column = self.config.columns.get("password_hash", "password_hash")

        cursor = self.conn.cursor()
        cursor.execute(
            f"SELECT {password_column} FROM {self.config.users_table} "
            f"WHERE {password_column} IS NOT NULL LIMIT {sample_size}"
        )
        rows = cursor.fetchall()
        cursor.close()

        if not rows:
            return "unknown"

        hashes = [row[0] for row in rows if row[0]]

        if not hashes:
            return "unknown"

        algorithms = []
        for hash_value in hashes:
            if isinstance(hash_value, bytes):
                hash_value = hash_value.decode("utf-8", errors="ignore")

            algorithm = self._detect_single_algorithm(hash_value)
            algorithms.append(algorithm)

        return max(set(algorithms), key=algorithms.count)

    def _detect_single_algorithm(self, hash_value: str) -> str:
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

    def close(self):
        if self.conn:
            self.conn.close()
            self.conn = None
