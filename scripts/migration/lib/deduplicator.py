"""User deduplication logic."""

from __future__ import annotations

from typing import Iterator

from . import ConflictRecord, DedupConfig, DedupStats, DedupStrategy, SourceUser


class DuplicateUserError(Exception):
    pass


class UserDeduplicator:
    def __init__(self, config: DedupConfig):
        self.config = config
        self.conflicts: list[ConflictRecord] = []
        self.stats = DedupStats()

    def deduplicate(self, users: Iterator[SourceUser]) -> list[SourceUser]:
        if self.config.strategy == DedupStrategy.NONE:
            result = list(users)
            self.stats.total = len(result)
            self.stats.unique = len(result)
            return result

        seen: dict[str, SourceUser] = {}
        unique_users: list[SourceUser] = []

        for user in users:
            self.stats.total += 1

            key = self._extract_key(user)

            if key is None:
                self.stats.skipped_no_key += 1
                continue

            if key not in seen:
                seen[key] = user
                unique_users.append(user)
                self.stats.unique += 1
            else:
                self.stats.duplicates += 1
                existing = seen[key]
                conflict = ConflictRecord(key=key, existing=existing, duplicate=user)
                self.conflicts.append(conflict)

                if self.config.strategy == DedupStrategy.ERROR:
                    raise DuplicateUserError(
                        f"Duplicate user found for key '{key}': "
                        f"existing_id={existing.id}, duplicate_id={user.id}"
                    )

                elif self.config.strategy == DedupStrategy.KEEP_LATEST:
                    if self._should_replace(user, existing):
                        for i, u in enumerate(unique_users):
                            if u is existing:
                                unique_users[i] = user
                                seen[key] = user
                                break

        return unique_users

    def _extract_key(self, user: SourceUser) -> str | None:
        if self.config.key == "email":
            return user.email.lower().strip() if user.email else None

        elif self.config.key == "phone":
            return user.phone.strip() if user.phone else None

        elif self.config.key == "username":
            return user.username.lower().strip() if user.username else None

        elif self.config.key == "email_or_phone":
            email_key = user.email.lower().strip() if user.email else None
            phone_key = user.phone.strip() if user.phone else None
            return email_key or phone_key

        elif self.config.key == "username_or_email":
            username_key = user.username.lower().strip() if user.username else None
            email_key = user.email.lower().strip() if user.email else None
            return username_key or email_key

        return None

    def _should_replace(self, new_user: SourceUser, existing_user: SourceUser) -> bool:
        if new_user.created_at is None:
            return False

        if existing_user.created_at is None:
            return True

        return new_user.created_at > existing_user.created_at
