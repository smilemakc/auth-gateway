#!/usr/bin/env python3
"""Auth Gateway User Migration Tool."""

import asyncio
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

import click
from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskID
from rich.table import Table

from lib import (
    load_config,
    MigrationConfig,
    MigrationReport,
    ImportStats,
    DedupStats,
    ConflictStrategy,
)
from lib.exporter import UserExporter
from lib.deduplicator import UserDeduplicator
from lib.importer import AuthGatewayImporter
from lib.verifier import MigrationVerifier
from lib.shadow import ShadowTableTransformer
from lib.report import MigrationReportGenerator

console = Console()


@click.group()
def cli():
    """Auth Gateway User Migration Tool"""
    pass


@cli.command()
@click.option('--config', '-c', default='config.yaml', help='Config file path')
def analyze(config: str):
    """Analyze source database without making changes."""
    try:
        cfg = load_config(config)
    except Exception as e:
        console.print(f"[red]Error loading config:[/red] {e}")
        sys.exit(1)

    console.print(Panel.fit("Analyzing Source Database", border_style="blue"))

    exporter: Optional[UserExporter] = None
    try:
        exporter = UserExporter(cfg.source)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            progress.add_task("Connecting to source database...", total=None)
            total_users = exporter.count_users()
            id_type = exporter.detect_id_type()
            password_algorithm = exporter.detect_password_algorithm()

        table = Table(title="Source Database Analysis", border_style="cyan")
        table.add_column("Metric", style="bold")
        table.add_column("Value", style="green")

        table.add_row("Total Users", str(total_users))
        table.add_row("ID Type", id_type or "unknown")
        table.add_row("Password Algorithm", password_algorithm or "unknown")
        table.add_row("Database", f"{cfg.source.host}:{cfg.source.port}/{cfg.source.database}")
        table.add_row("Table", cfg.source.users_table)

        console.print(table)

        console.print("\n[yellow]Checking for duplicates...[/yellow]")
        users_iter = exporter.export_users()
        deduplicator = UserDeduplicator(cfg.dedup)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            progress.add_task("Scanning for duplicate records...", total=None)
            deduplicated = deduplicator.deduplicate(users_iter)

        conflicts = deduplicator.get_conflicts()

        dedup_table = Table(title="Deduplication Preview", border_style="yellow")
        dedup_table.add_column("Metric", style="bold")
        dedup_table.add_column("Value")

        dedup_table.add_row("Unique users", str(len(deduplicated)), style="green")
        dedup_table.add_row("Duplicate records", str(len(conflicts)), style="red" if conflicts else "green")

        if conflicts:
            dedup_table.add_row("Dedup strategy", cfg.dedup.strategy.value, style="cyan")

        console.print(dedup_table)

        if conflicts and len(conflicts) <= 10:
            console.print("\n[yellow]Sample conflicts:[/yellow]")
            conflict_table = Table(border_style="red")
            conflict_table.add_column("Key")
            conflict_table.add_column("Existing ID")
            conflict_table.add_column("Duplicate ID")

            for conf in conflicts[:10]:
                conflict_table.add_row(
                    conf.key,
                    conf.existing.id,
                    conf.duplicate.id,
                )

            console.print(conflict_table)
        elif len(conflicts) > 10:
            console.print(f"\n[yellow]Too many conflicts to display ({len(conflicts)} total)[/yellow]")

        console.print(Panel.fit(
            "[green]Analysis complete.[/green] No changes were made to any database.",
            border_style="green"
        ))

    except Exception as e:
        console.print(f"\n[red]Error during analysis:[/red] {e}")
        sys.exit(1)
    finally:
        if exporter:
            exporter.close()


@cli.command()
@click.option('--config', '-c', default='config.yaml', help='Config file path')
@click.option('--dry-run/--no-dry-run', default=True, help='Dry run mode (no actual changes)')
@click.option(
    '--step',
    type=click.Choice(['import', 'verify', 'shadow', 'all']),
    default='all',
    help='Migration step to execute'
)
def migrate(config: str, dry_run: bool, step: str):
    """Run migration (full or step-by-step)."""
    try:
        cfg = load_config(config)
    except Exception as e:
        console.print(f"[red]Error loading config:[/red] {e}")
        sys.exit(1)

    mode_text = "[yellow]DRY RUN[/yellow]" if dry_run else "[red]LIVE MODE[/red]"
    console.print(Panel.fit(
        f"Starting Migration\nMode: {mode_text}\nStep: [cyan]{step}[/cyan]",
        border_style="blue"
    ))

    if not dry_run:
        console.print("[red bold]WARNING: This will modify the target database![/red bold]")
        if not click.confirm("Continue?"):
            console.print("Migration cancelled.")
            sys.exit(0)

    report = MigrationReport(started_at=datetime.now())

    try:
        if step in ('import', 'all'):
            _run_import_step(cfg, dry_run, report)

        if step in ('verify', 'all'):
            asyncio.run(_run_verify_step(cfg, report))

        if step in ('shadow', 'all') and cfg.shadow.enabled and not dry_run:
            _run_shadow_step(cfg, report)

        report.finished_at = datetime.now()

        console.print(Panel.fit(
            f"[green]Migration completed in {report.duration_seconds:.2f}s[/green]",
            border_style="green"
        ))

        _print_migration_summary(report)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        report_path = Path(f"migration_report_{timestamp}.json")
        report_gen = MigrationReportGenerator(report)
        report_gen.save(report_path)
        console.print(f"\n[cyan]Report saved to:[/cyan] {report_path}")

    except Exception as e:
        report.errors.append(str(e))
        report.finished_at = datetime.now()
        console.print(f"\n[red]Migration failed:[/red] {e}")
        sys.exit(1)


def _run_import_step(cfg: MigrationConfig, dry_run: bool, report: MigrationReport):
    """Execute import step: export → dedup → import."""
    console.print("\n[bold blue]Step 1: Export & Deduplicate[/bold blue]")

    exporter: Optional[UserExporter] = None
    try:
        exporter = UserExporter(cfg.source)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
            console=console,
        ) as progress:
            export_task = progress.add_task("Exporting users...", total=exporter.count_users())
            users_iter = exporter.export_users()

            users_list = []
            for user in users_iter:
                users_list.append(user)
                progress.update(export_task, advance=1)

        deduplicator = UserDeduplicator(cfg.dedup)
        deduplicated = deduplicator.deduplicate(iter(users_list))

        report.dedup_stats = DedupStats(
            total=len(users_list),
            unique=len(deduplicated),
            duplicates=len(users_list) - len(deduplicated),
        )

        console.print(f"[green]Exported {len(users_list)} users, deduplicated to {len(deduplicated)}[/green]")

        console.print("\n[bold blue]Step 2: Import to Auth Gateway[/bold blue]")

        if dry_run:
            console.print("[yellow]Dry run: skipping actual import[/yellow]")
            report.import_stats = ImportStats(total=len(deduplicated))
        else:
            asyncio.run(_run_import_async(cfg, deduplicated, report))

    finally:
        if exporter:
            exporter.close()


async def _run_import_async(cfg: MigrationConfig, users: list, report: MigrationReport):
    """Execute async import."""
    importer = AuthGatewayImporter(
        cfg.target,
        cfg.password,
        cfg.conflicts,
        cfg.workers
    )

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
        console=console,
    ) as progress:
        import_task = progress.add_task("Importing users...", total=len(users))

        results = await importer.import_users(users)
        progress.update(import_task, completed=len(users))

    report.import_stats = ImportStats(
        total=len(results),
        created=sum(1 for r in results if r.status.value == "created"),
        existing=sum(1 for r in results if r.status.value == "existing"),
        skipped=sum(1 for r in results if r.status.value == "skipped"),
        updated=sum(1 for r in results if r.status.value == "updated"),
        errors=sum(1 for r in results if r.status.value == "error"),
    )

    console.print(f"[green]Import complete: {report.import_stats.created} created, "
                  f"{report.import_stats.errors} errors[/green]")


async def _run_verify_step(cfg: MigrationConfig, report: MigrationReport):
    """Execute verification step."""
    console.print("\n[bold blue]Step 3: Verify Migration[/bold blue]")

    verifier = MigrationVerifier(cfg.source, cfg.target)

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
    ) as progress:
        progress.add_task("Verifying migration...", total=None)
        verification_report = await verifier.verify()

    report.verification = verification_report

    if verification_report.passed:
        console.print("[green]Verification passed![/green]")
    else:
        console.print("[red]Verification failed![/red]")
        _print_verification_details(verification_report)


def _run_shadow_step(cfg: MigrationConfig, report: MigrationReport):
    """Execute shadow table transformation."""
    console.print("\n[bold blue]Step 4: Shadow Table Transformation[/bold blue]")

    transformer = ShadowTableTransformer(cfg.shadow)

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
    ) as progress:
        progress.add_task("Applying shadow transformation...", total=None)
        sql = transformer.generate_sql()

    console.print(f"[green]Generated {len(sql)} SQL statements[/green]")

    from lib import ShadowReport
    report.shadow = ShadowReport(sql=sql, applied=False)


@cli.command()
@click.option('--config', '-c', default='config.yaml', help='Config file path')
def verify(config: str):
    """Verify migration integrity."""
    try:
        cfg = load_config(config)
    except Exception as e:
        console.print(f"[red]Error loading config:[/red] {e}")
        sys.exit(1)

    console.print(Panel.fit("Verifying Migration", border_style="blue"))

    async def run_verification():
        verifier = MigrationVerifier(cfg.source, cfg.target)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            progress.add_task("Verifying migration...", total=None)
            return await verifier.verify()

    try:
        verification_report = asyncio.run(run_verification())

        if verification_report.passed:
            console.print(Panel.fit(
                "[green bold]Verification PASSED[/green bold]",
                border_style="green"
            ))
        else:
            console.print(Panel.fit(
                "[red bold]Verification FAILED[/red bold]",
                border_style="red"
            ))

        _print_verification_details(verification_report)

        if not verification_report.passed:
            sys.exit(1)

    except Exception as e:
        console.print(f"\n[red]Verification error:[/red] {e}")
        sys.exit(1)


@cli.command('generate-shadow-sql')
@click.option('--config', '-c', default='config.yaml', help='Config file path')
@click.option('--output', '-o', default='shadow_migration.sql', help='Output SQL file')
def generate_shadow_sql(config: str, output: str):
    """Generate SQL migration file for shadow table conversion."""
    try:
        cfg = load_config(config)
    except Exception as e:
        console.print(f"[red]Error loading config:[/red] {e}")
        sys.exit(1)

    if not cfg.shadow.enabled:
        console.print("[yellow]Shadow mode is disabled in config[/yellow]")
        sys.exit(0)

    console.print(Panel.fit("Generating Shadow Table Migration SQL", border_style="blue"))

    try:
        transformer = ShadowTableTransformer(cfg.shadow)
        output_path = Path(output)

        transformer.generate_migration_file(output_path)

        console.print(Panel.fit(
            f"[green]SQL migration file generated:[/green]\n{output_path.absolute()}",
            border_style="green"
        ))

    except Exception as e:
        console.print(f"\n[red]Error generating SQL:[/red] {e}")
        sys.exit(1)


def _print_migration_summary(report: MigrationReport):
    """Print summary table of migration results."""
    table = Table(title="Migration Summary", border_style="cyan", show_header=False)
    table.add_column("Metric", style="bold")
    table.add_column("Value")

    table.add_row("Duration", f"{report.duration_seconds:.2f}s")

    if report.dedup_stats.total > 0:
        table.add_row("Total users", str(report.dedup_stats.total))
        table.add_row("Unique users", str(report.dedup_stats.unique), style="green")
        table.add_row("Duplicates removed", str(report.dedup_stats.duplicates), style="yellow")

    if report.import_stats.total > 0:
        table.add_row("", "")
        table.add_row("Users imported", str(report.import_stats.total))
        table.add_row("Created", str(report.import_stats.created), style="green")
        table.add_row("Updated", str(report.import_stats.updated), style="cyan")
        table.add_row("Skipped", str(report.import_stats.skipped), style="yellow")
        table.add_row("Errors", str(report.import_stats.errors), style="red")

    if report.verification:
        table.add_row("", "")
        status = "[green]PASSED[/green]" if report.verification.passed else "[red]FAILED[/red]"
        table.add_row("Verification", status)

    console.print(table)


def _print_verification_details(verification_report):
    """Print detailed verification results."""
    checks_table = Table(title="Verification Checks", border_style="cyan")
    checks_table.add_column("Check", style="bold")
    checks_table.add_column("Expected")
    checks_table.add_column("Actual")
    checks_table.add_column("Status")

    for check in verification_report.checks:
        status = "[green]✓[/green]" if check.passed else "[red]✗[/red]"
        checks_table.add_row(
            check.name,
            str(check.expected),
            str(check.actual),
            status
        )

    console.print(checks_table)

    if verification_report.missing_users:
        console.print(f"\n[red]Missing users: {len(verification_report.missing_users)}[/red]")
        if len(verification_report.missing_users) <= 10:
            missing_table = Table(border_style="red")
            missing_table.add_column("User ID")
            missing_table.add_column("Email")

            for user in verification_report.missing_users[:10]:
                missing_table.add_row(user["user_id"], user["email"])

            console.print(missing_table)

    if verification_report.mismatches:
        console.print(f"\n[yellow]Field mismatches: {len(verification_report.mismatches)}[/yellow]")
        if len(verification_report.mismatches) <= 10:
            mismatch_table = Table(border_style="yellow")
            mismatch_table.add_column("User ID")
            mismatch_table.add_column("Field")
            mismatch_table.add_column("Expected")
            mismatch_table.add_column("Actual")

            for mismatch in verification_report.mismatches[:10]:
                mismatch_table.add_row(
                    mismatch["user_id"],
                    mismatch["field"],
                    mismatch["expected"],
                    mismatch["actual"]
                )

            console.print(mismatch_table)


if __name__ == '__main__':
    cli()
