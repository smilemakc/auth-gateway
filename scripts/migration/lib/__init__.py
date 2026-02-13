"""Auth Gateway User Migration Library — data models and configuration."""

from __future__ import annotations

import os
import re
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Optional

import yaml


# ──────────────────────────────────────────────
# Enums
# ──────────────────────────────────────────────

class IdStrategy(str, Enum):
    PRESERVE_UUID = "preserve_uuid"
    GENERATE_NEW = "generate_new"


class PasswordStrategy(str, Enum):
    TRANSFER = "transfer"
    FORCE_RESET = "force_reset"
    NONE = "none"


class ConflictStrategy(str, Enum):
    SKIP = "skip"
    UPDATE = "update"
    ERROR = "error"


class DedupStrategy(str, Enum):
    NONE = "none"
    KEEP_LATEST = "keep_latest"
    KEEP_FIRST = "keep_first"
    ERROR = "error"


class ImportStatus(str, Enum):
    CREATED = "created"
    EXISTING = "existing"
    SKIPPED = "skipped"
    UPDATED = "updated"
    ERROR = "error"


# ──────────────────────────────────────────────
# Configuration dataclasses
# ──────────────────────────────────────────────

@dataclass
class SourceConfig:
    type: str = "postgresql"
    host: str = "localhost"
    port: int = 5432
    database: str = ""
    user: str = ""
    password: str = ""
    ssl: bool = False
    users_table: str = "users"
    id_strategy: IdStrategy = IdStrategy.PRESERVE_UUID
    columns: dict[str, str] = field(default_factory=lambda: {
        "id": "id",
        "email": "email",
        "username": "username",
        "full_name": "full_name",
        "phone": "phone",
        "is_active": "is_active",
        "email_verified": "email_verified",
        "created_at": "created_at",
        # Optional — include only if source DB has password column:
        # "password_hash": "password_hash",
    })
    batch_size: int = 100


@dataclass
class TargetConfig:
    base_url: str = "http://localhost:3000"
    api_key: str = ""
    application_id: str = ""


@dataclass
class PasswordConfig:
    strategy: PasswordStrategy = PasswordStrategy.TRANSFER
    source_algorithm: str = "bcrypt"


@dataclass
class DedupConfig:
    key: str = "email"
    strategy: DedupStrategy = DedupStrategy.KEEP_LATEST


@dataclass
class ValidationConfig:
    skip_email_validation: bool = False
    skip_phone_validation: bool = False


@dataclass
class RolesConfig:
    enabled: bool = False
    source_column: str = "role_id"
    source_table: str = ""
    source_user_id_column: str = "user_id"
    source_role_id_column: str = "role_id"
    mapping: dict[int, str] = field(default_factory=dict)


@dataclass
class ShadowColumnConfig:
    name: str
    type: str
    default: str


@dataclass
class ShadowConfig:
    enabled: bool = True
    users_table: str = "users"
    drop_columns: list[str] = field(default_factory=lambda: ["password_hash"])
    add_columns: list[ShadowColumnConfig] = field(default_factory=list)


@dataclass
class MigrationConfig:
    mode: str = "dry-run"
    batch_size: int = 100
    workers: int = 4
    source: SourceConfig = field(default_factory=SourceConfig)
    target: TargetConfig = field(default_factory=TargetConfig)
    password: PasswordConfig = field(default_factory=PasswordConfig)
    conflicts: ConflictStrategy = ConflictStrategy.SKIP
    dedup: DedupConfig = field(default_factory=DedupConfig)
    shadow: ShadowConfig = field(default_factory=ShadowConfig)
    validation: ValidationConfig = field(default_factory=ValidationConfig)
    roles: RolesConfig = field(default_factory=RolesConfig)


# ──────────────────────────────────────────────
# Domain models
# ──────────────────────────────────────────────

@dataclass
class SourceUser:
    id: str
    email: Optional[str] = None
    username: Optional[str] = None
    password_hash: Optional[str] = None
    full_name: Optional[str] = None
    phone: Optional[str] = None
    is_active: bool = True
    email_verified: bool = False
    phone_verified: bool = False
    created_at: Optional[datetime] = None
    roles: list[int] = field(default_factory=list)


@dataclass
class ImportResult:
    source_id: str
    ag_id: Optional[str] = None
    status: ImportStatus = ImportStatus.CREATED
    note: str = ""


@dataclass
class ConflictRecord:
    key: str
    existing: SourceUser
    duplicate: SourceUser


# ──────────────────────────────────────────────
# Statistics
# ──────────────────────────────────────────────

@dataclass
class DedupStats:
    total: int = 0
    unique: int = 0
    duplicates: int = 0
    skipped_no_key: int = 0


@dataclass
class ImportStats:
    total: int = 0
    created: int = 0
    existing: int = 0
    skipped: int = 0
    updated: int = 0
    errors: int = 0


@dataclass
class VerificationCheck:
    name: str
    expected: Any
    actual: Any
    passed: bool = False

    def __post_init__(self):
        self.passed = self.expected == self.actual


@dataclass
class VerificationReport:
    checks: list[VerificationCheck] = field(default_factory=list)
    missing_users: list[dict[str, str]] = field(default_factory=list)
    mismatches: list[dict[str, str]] = field(default_factory=list)

    def add_check(self, name: str, expected: Any, actual: Any):
        self.checks.append(VerificationCheck(name=name, expected=expected, actual=actual))

    def add_missing(self, user_id: str, email: str):
        self.missing_users.append({"user_id": user_id, "email": email})

    def add_mismatch(self, user_id: str, field_name: str, expected: str, actual: str):
        self.mismatches.append({
            "user_id": user_id,
            "field": field_name,
            "expected": expected,
            "actual": actual,
        })

    @property
    def passed(self) -> bool:
        return (
            all(c.passed for c in self.checks)
            and not self.missing_users
            and not self.mismatches
        )


@dataclass
class ShadowReport:
    sql: list[str] = field(default_factory=list)
    applied: bool = False


@dataclass
class MigrationReport:
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    dedup_stats: DedupStats = field(default_factory=DedupStats)
    import_stats: ImportStats = field(default_factory=ImportStats)
    verification: Optional[VerificationReport] = None
    shadow: Optional[ShadowReport] = None
    errors: list[str] = field(default_factory=list)

    @property
    def duration_seconds(self) -> float:
        if self.started_at and self.finished_at:
            return (self.finished_at - self.started_at).total_seconds()
        return 0.0


# ──────────────────────────────────────────────
# Config loader
# ──────────────────────────────────────────────

_ENV_VAR_RE = re.compile(r"\$\{(\w+)}")


def _resolve_env(value: str) -> str:
    """Replace ${VAR} with environment variable value."""
    def _replace(m: re.Match) -> str:
        return os.environ.get(m.group(1), m.group(0))
    return _ENV_VAR_RE.sub(_replace, value)


def load_config(path: str | Path) -> MigrationConfig:
    """Load and parse migration config from YAML file."""
    with open(path) as f:
        raw = yaml.safe_load(f)

    migration = raw.get("migration", {})
    src = raw.get("source", {})
    tgt = raw.get("target", {})
    pwd = raw.get("password", {})
    conflicts = raw.get("conflicts", {})
    dedup = raw.get("deduplication", {})
    shadow_raw = raw.get("shadow", {})
    validation_raw = raw.get("validation", {})
    roles_raw = raw.get("roles", {})

    source = SourceConfig(
        type=src.get("type", "postgresql"),
        host=src.get("host", "localhost"),
        port=src.get("port", 5432),
        database=src.get("database", ""),
        user=src.get("user", ""),
        password=_resolve_env(src.get("password", "")),
        ssl=src.get("ssl", False),
        users_table=src.get("users_table", "users"),
        id_strategy=IdStrategy(src.get("id_strategy", "preserve_uuid")),
        columns=src.get("columns", {}),
        batch_size=migration.get("batch_size", 100),
    )

    target = TargetConfig(
        base_url=tgt.get("base_url", "http://localhost:3000"),
        api_key=_resolve_env(tgt.get("api_key", "")),
        application_id=tgt.get("application_id", ""),
    )

    password = PasswordConfig(
        strategy=PasswordStrategy(pwd.get("strategy", "transfer")),
        source_algorithm=pwd.get("source_algorithm", "bcrypt"),
    )

    shadow_add_cols = [
        ShadowColumnConfig(name=c["name"], type=c["type"], default=c.get("default", "NULL"))
        for c in shadow_raw.get("add_columns", [])
    ]
    shadow = ShadowConfig(
        enabled=shadow_raw.get("enabled", True),
        users_table=src.get("users_table", "users"),
        drop_columns=shadow_raw.get("drop_columns", ["password_hash"]),
        add_columns=shadow_add_cols,
    )

    validation = ValidationConfig(
        skip_email_validation=validation_raw.get("skip_email_validation", False),
        skip_phone_validation=validation_raw.get("skip_phone_validation", False),
    )

    roles_mapping_raw = roles_raw.get("mapping", {})
    if not isinstance(roles_mapping_raw, dict):
        roles_mapping_raw = {}
    roles_mapping: dict[int, str] = {}
    for k, v in roles_mapping_raw.items():
        if isinstance(v, list):
            # mapping: {1: ["admin", "moderator"]} → multiple role names per source ID
            for role_name in v:
                roles_mapping[int(k)] = str(role_name)
        else:
            roles_mapping[int(k)] = str(v)
    roles = RolesConfig(
        enabled=roles_raw.get("enabled", False),
        source_column=roles_raw.get("source_column", "role_id"),
        source_table=roles_raw.get("source_table", ""),
        source_user_id_column=roles_raw.get("source_user_id_column", "user_id"),
        source_role_id_column=roles_raw.get("source_role_id_column", "role_id"),
        mapping=roles_mapping,
    )

    return MigrationConfig(
        mode=migration.get("mode", "dry-run"),
        batch_size=migration.get("batch_size", 100),
        workers=migration.get("workers", 4),
        source=source,
        target=target,
        password=password,
        conflicts=ConflictStrategy(conflicts.get("strategy", "skip")),
        dedup=DedupConfig(
            key=dedup.get("key", "email"),
            strategy=DedupStrategy(dedup.get("strategy", "keep_latest")),
        ),
        shadow=shadow,
        validation=validation,
        roles=roles,
    )
