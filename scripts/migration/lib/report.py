"""Migration report generation - JSON and console output."""

from __future__ import annotations

import json
from dataclasses import asdict
from datetime import datetime
from pathlib import Path

from rich.console import Console
from rich.table import Table

from . import MigrationReport


def _default_serializer(obj):
    if isinstance(obj, datetime):
        return obj.isoformat()
    raise TypeError(f"Object of type {type(obj).__name__} is not JSON serializable")


class MigrationReportGenerator:
    def __init__(self, report: MigrationReport):
        self.report = report

    def to_json(self) -> str:
        data = asdict(self.report)
        return json.dumps(data, indent=2, default=_default_serializer)

    def save(self, path: str | Path):
        path_obj = Path(path)
        path_obj.parent.mkdir(parents=True, exist_ok=True)
        path_obj.write_text(self.to_json())

    def print_summary(self):
        console = Console()

        console.print("\n[bold cyan]Migration Report[/bold cyan]\n")

        info_table = Table(show_header=False, box=None)
        info_table.add_column("Field", style="bold")
        info_table.add_column("Value")

        started = (
            self.report.started_at.strftime("%Y-%m-%d %H:%M:%S")
            if self.report.started_at
            else "N/A"
        )
        finished = (
            self.report.finished_at.strftime("%Y-%m-%d %H:%M:%S")
            if self.report.finished_at
            else "N/A"
        )
        duration = f"{self.report.duration_seconds:.2f}s"

        info_table.add_row("Started", started)
        info_table.add_row("Finished", finished)
        info_table.add_row("Duration", duration)

        console.print(info_table)
        console.print()

        dedup_table = Table(title="Deduplication Statistics")
        dedup_table.add_column("Metric", style="cyan")
        dedup_table.add_column("Count", justify="right", style="green")

        dedup = self.report.dedup_stats
        dedup_table.add_row("Total", str(dedup.total))
        dedup_table.add_row("Unique", str(dedup.unique))
        dedup_table.add_row("Duplicates", str(dedup.duplicates))
        dedup_table.add_row("Skipped (no key)", str(dedup.skipped_no_key))

        console.print(dedup_table)
        console.print()

        import_table = Table(title="Import Statistics")
        import_table.add_column("Status", style="cyan")
        import_table.add_column("Count", justify="right", style="green")

        imp = self.report.import_stats
        import_table.add_row("Total", str(imp.total))
        import_table.add_row("Created", str(imp.created))
        import_table.add_row("Existing", str(imp.existing))
        import_table.add_row("Skipped", str(imp.skipped))
        import_table.add_row("Updated", str(imp.updated))
        import_table.add_row("Errors", f"[red]{imp.errors}[/red]" if imp.errors else str(imp.errors))

        console.print(import_table)
        console.print()

        if self.report.verification:
            ver = self.report.verification
            ver_table = Table(title="Verification Results")
            ver_table.add_column("Check", style="cyan")
            ver_table.add_column("Expected", justify="right")
            ver_table.add_column("Actual", justify="right")
            ver_table.add_column("Status", justify="center")

            for check in ver.checks:
                status = "[green]✓[/green]" if check.passed else "[red]✗[/red]"
                ver_table.add_row(
                    check.name,
                    str(check.expected),
                    str(check.actual),
                    status,
                )

            console.print(ver_table)
            console.print()

            if ver.missing_users:
                console.print(f"[yellow]Missing users: {len(ver.missing_users)}[/yellow]")
                for missing in ver.missing_users[:10]:
                    console.print(
                        f"  - {missing['user_id']} ({missing['email']})", style="dim"
                    )
                if len(ver.missing_users) > 10:
                    console.print(f"  ... and {len(ver.missing_users) - 10} more", style="dim")
                console.print()

            if ver.mismatches:
                console.print(f"[yellow]Field mismatches: {len(ver.mismatches)}[/yellow]")
                for mismatch in ver.mismatches[:10]:
                    console.print(
                        f"  - {mismatch['user_id']}: {mismatch['field']} "
                        f"(expected: {mismatch['expected']}, actual: {mismatch['actual']})",
                        style="dim",
                    )
                if len(ver.mismatches) > 10:
                    console.print(f"  ... and {len(ver.mismatches) - 10} more", style="dim")
                console.print()

            overall_status = "[green]PASSED[/green]" if ver.passed else "[red]FAILED[/red]"
            console.print(f"Overall verification: {overall_status}\n")

        if self.report.shadow:
            shadow = self.report.shadow
            console.print("[bold]Shadow Mode SQL[/bold]")
            for sql in shadow.sql:
                console.print(f"  {sql}", style="dim")
            console.print(f"Applied: {shadow.applied}\n")

        if self.report.errors:
            console.print(f"[red bold]Errors: {len(self.report.errors)}[/red bold]")
            for error in self.report.errors[:10]:
                console.print(f"  - {error}", style="red dim")
            if len(self.report.errors) > 10:
                console.print(f"  ... and {len(self.report.errors) - 10} more", style="red dim")
            console.print()
