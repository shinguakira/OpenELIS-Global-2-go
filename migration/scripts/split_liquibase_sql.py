#!/usr/bin/env python3
"""Split a Liquibase `updateSql`/offline-mode SQL dump into numbered goose
migration files.

Input is the file produced by (see migration/liquibase-to-goose-plan.md § 2):

    liquibase --classpath=src/main/resources \
              --changelog-file=liquibase/base-changelog.xml \
              --output-file=<full.sql> \
              updateSQL --url=offline:postgresql

Each changeset in that file is preceded by a marker line of the exact form:

    -- Changeset <changelog/relative/path.xml>::<id>::<author>

This script splits on that marker, and for every changeset emits one file
    db/migrations/<seq>_<slug>.sql
in goose format (`-- +goose Up` / `-- +goose Down`).

Down-block generation is a best-effort heuristic (see `infer_down`) for the
small set of single-statement, unambiguous patterns (CREATE TABLE, CREATE
SEQUENCE, CREATE INDEX, single-column ADD COLUMN, CREATE VIEW). Everything
else — multi-statement bodies, DML, ALTER COLUMN TYPE, DROP, DO blocks, added
constraints — gets an honest `-- TODO` placeholder instead of a guessed
rollback. See migration/liquibase-to-goose-plan.md § 7 "Risk items" for why:
guessing wrong here is worse than admitting the gap.

Usage:
    python3 split_liquibase_sql.py <full.sql> <out_dir>
"""

import re
import sys
from pathlib import Path

MARKER_RE = re.compile(r"^-- Changeset (?P<path>\S+)::(?P<id>[^:]+)::(?P<author>.+)$")

# Statements needing NO TRANSACTION in Postgres (can't run inside a tx block).
NO_TXN_RE = re.compile(
    r"CREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY|DROP\s+INDEX\s+CONCURRENTLY|"
    r"ALTER\s+TYPE\s+\S+\s+ADD\s+VALUE",
    re.IGNORECASE,
)

# (pattern, replacement-template) — applied only when the ENTIRE up-body
# (after stripping leading `-- comment` lines) is a single statement matching
# the pattern. Order matters: first match wins.
DOWN_PATTERNS = [
    (
        re.compile(
            r"^CREATE TABLE\s+(?:IF NOT EXISTS\s+)?([\w.]+)\s*\(.*\)\s*;?\s*$",
            re.IGNORECASE | re.DOTALL,
        ),
        "DROP TABLE IF EXISTS {0};",
    ),
    (
        re.compile(
            r"^CREATE\s+(?:OR REPLACE\s+)?VIEW\s+([\w.]+)\s+AS\b.*$",
            re.IGNORECASE | re.DOTALL,
        ),
        "DROP VIEW IF EXISTS {0};",
    ),
    (
        re.compile(
            r"^CREATE\s+SEQUENCE\s+(?:IF NOT EXISTS\s+)?([\w.]+).*$",
            re.IGNORECASE | re.DOTALL,
        ),
        "DROP SEQUENCE IF EXISTS {0};",
    ),
    (
        re.compile(
            r"^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:CONCURRENTLY\s+)?"
            r"(?:IF NOT EXISTS\s+)?([\w.\"]+)\s+ON\b.*$",
            re.IGNORECASE | re.DOTALL,
        ),
        "DROP INDEX IF EXISTS {0};",
    ),
    (
        re.compile(
            r"^ALTER TABLE\s+([\w.]+)\s+ADD\s+(?:COLUMN\s+)?(?:IF NOT EXISTS\s+)?"
            r"(\"?\w+\"?)\s+[\w].*$",
            re.IGNORECASE | re.DOTALL,
        ),
        "ALTER TABLE {0} DROP COLUMN IF EXISTS {1};",
    ),
]

# Statements containing any of these anywhere make a body ineligible for
# single-statement down inference even if the leading pattern matched
# (e.g. a CREATE TABLE followed by extra trailing statements).
_MULTI_STATEMENT_GUARD = re.compile(r";\s*\S")  # a ';' followed by more non-ws content

# --- idempotency guards -----------------------------------------------------
# Confirmed empirically (running the generated migrations against a clean DB,
# see migration/liquibase-to-goose-plan.md): plain extracted SQL hits real
# "duplicate key value violates unique constraint" errors on ordinary
# INSERTs. Root cause: 218 of the source XML files use <preConditions
# onFail="MARK_RAN">, which Liquibase evaluates against live DB state at
# apply time (skip silently if the precondition fails) — offline-mode
# extraction has no DB to check, so it can't reproduce that skip and just
# always emits the statement. These rewrites make the safely-rewritable
# statements idempotent (CREATE TABLE/SEQUENCE/INDEX -> IF NOT EXISTS,
# ADD COLUMN -> IF NOT EXISTS, plain INSERT -> ON CONFLICT DO NOTHING), the
# same practical effect MARK_RAN had — applied per-statement within a
# changeset (see split_top_level_statements), not just to whole-body
# single-statement changesets, so a multi-statement changeset with one
# guardable CREATE TABLE and one un-guardable UPDATE still gets the former
# guarded. Statement shapes outside the pattern list (UPDATE, DELETE, ALTER
# COLUMN TYPE, DROP, DO blocks, added constraints) are left untouched — the
# manifest records which changesets got zero guards so those are an
# itemized follow-up, not a silent gap.
# NOTE on all four "insert IF NOT EXISTS" keyword regexes below: matching
# "is this already guarded" as a negative lookahead glued directly onto a
# variable-length \s+ is a real bug, confirmed empirically — some of this
# changelog's own SQL already has irregular spacing ("CREATE SEQUENCE  IF
# NOT EXISTS ...", double space). A combined `\s+(?!IF NOT EXISTS\b)` regex
# lets the engine BACKTRACK \s+ down to consuming just one of the two
# spaces so the lookahead technically finds a non-match (" IF NOT EXISTS"
# with a leading space != "IF NOT EXISTS") and wrongly reports "not yet
# guarded" — producing "CREATE SEQUENCE IF NOT EXISTS  IF NOT EXISTS ...".
# Fixed by never combining them: match only the keyword prefix, then check
# "already guarded" as an independent regex against whatever follows,
# tolerant of any amount of whitespace.
_CREATE_TABLE_KW = re.compile(r"^CREATE TABLE\s+", re.IGNORECASE)
_CREATE_SEQ_KW = re.compile(r"^CREATE SEQUENCE\s+", re.IGNORECASE)
_CREATE_INDEX_KW = re.compile(r"^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:CONCURRENTLY\s+)?", re.IGNORECASE)
_ALREADY_IFNE_RE = re.compile(r"^\s*IF\s+NOT\s+EXISTS\b", re.IGNORECASE)

# DROP-side mirror: confirmed empirically (full-scan diagnostic pass) that
# a from-scratch apply can hit a changeset DROPping something an earlier
# changeset never created in THIS sequence (e.g. system_config, a column
# already gone) — same underlying onFail=MARK_RAN gap, opposite direction.
_DROP_TABLE_KW = re.compile(r"^DROP TABLE\s+", re.IGNORECASE)
_DROP_SEQ_KW = re.compile(r"^DROP SEQUENCE\s+", re.IGNORECASE)
_DROP_INDEX_KW = re.compile(r"^DROP\s+INDEX\s+(?:CONCURRENTLY\s+)?", re.IGNORECASE)
_DROP_VIEW_KW = re.compile(r"^DROP VIEW\s+", re.IGNORECASE)
_ALTER_DROP_COLUMN_KW = re.compile(r"^ALTER TABLE\s+[\w.]+\s+DROP\s+COLUMN\s+", re.IGNORECASE)
_ALREADY_IFE_RE = re.compile(r"^\s*IF\s+EXISTS\b", re.IGNORECASE)


def _insert_guard_after(stripped, keyword_re, already_re, guard_phrase):
    """If stripped starts with keyword_re and isn't already followed by
    guard_phrase (per already_re, tolerating any whitespace), return the
    rewritten string with guard_phrase inserted; else None."""
    m = keyword_re.match(stripped)
    if not m:
        return None
    if already_re.match(stripped[m.end() :]):
        return None
    return stripped[: m.end()] + guard_phrase + " " + stripped[m.end() :]


def _insert_if_not_exists_after(stripped, keyword_re):
    return _insert_guard_after(stripped, keyword_re, _ALREADY_IFNE_RE, "IF NOT EXISTS")


def _insert_if_exists_after(stripped, keyword_re):
    return _insert_guard_after(stripped, keyword_re, _ALREADY_IFE_RE, "IF EXISTS")


# "ADD" on ALTER TABLE is ambiguous — ADD COLUMN vs ADD CONSTRAINT/PRIMARY
# KEY/UNIQUE/FOREIGN KEY/CHECK/EXCLUDE, and only ADD [COLUMN] accepts IF NOT
# EXISTS in Postgres. Confirmed by hitting a real syntax error empirically
# ("ADD IF NOT EXISTS CONSTRAINT ... UNIQUE" is invalid) before this
# exclusion was added.
_ALTER_ADD_KW = re.compile(r"^ALTER TABLE\s+[\w.]+\s+ADD\s+(?:COLUMN\s+)?", re.IGNORECASE)
_ALTER_ADD_SUBCLAUSE_RE = re.compile(
    r"^\s*(?:CONSTRAINT|PRIMARY|UNIQUE|FOREIGN|CHECK|EXCLUDE)\b", re.IGNORECASE
)
_INSERT_INTO_RE = re.compile(r"^INSERT\s+INTO\b", re.IGNORECASE)

_DOLLAR_TAG_RE = re.compile(r"\$[A-Za-z_]*\$")


def split_top_level_statements(sql):
    """Split on ';' while respecting '...'/"..." string literals and
    $$.../$tag$...$tag$ dollar-quoting (so a DO $$ ... ; ... $$ block, whose
    body legitimately contains ';', is never split apart). Returns a list of
    statement strings with the separating ';' removed; a final statement
    with no trailing ';' is included as-is."""
    stmts = []
    buf = []
    i, n = 0, len(sql)
    dollar_tag = None
    in_squote = False
    in_dquote = False
    while i < n:
        c = sql[i]
        if dollar_tag:
            if sql.startswith(dollar_tag, i):
                buf.append(dollar_tag)
                i += len(dollar_tag)
                dollar_tag = None
            else:
                buf.append(c)
                i += 1
            continue
        if in_squote:
            buf.append(c)
            if c == "'":
                if sql[i + 1 : i + 2] == "'":  # doubled '' escape
                    buf.append("'")
                    i += 2
                    continue
                in_squote = False
            i += 1
            continue
        if in_dquote:
            buf.append(c)
            if c == '"':
                in_dquote = False
            i += 1
            continue
        if c == "'":
            in_squote = True
            buf.append(c)
            i += 1
            continue
        if c == '"':
            in_dquote = True
            buf.append(c)
            i += 1
            continue
        if c == "$":
            m = _DOLLAR_TAG_RE.match(sql, i)
            if m:
                dollar_tag = m.group(0)
                buf.append(dollar_tag)
                i = m.end()
                continue
        if c == ";":
            stmts.append("".join(buf))
            buf = []
            i += 1
            continue
        buf.append(c)
        i += 1
    tail = "".join(buf).strip()
    if tail:
        stmts.append(tail)
    return [s.strip() for s in stmts if s.strip()]


def _guard_single_statement(stripped):
    """Apply one idempotency rewrite to a single already-isolated statement
    (no trailing ';'). Returns (possibly-rewritten statement, applied?)."""
    rewritten = _insert_if_not_exists_after(stripped, _CREATE_TABLE_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_not_exists_after(stripped, _CREATE_SEQ_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_not_exists_after(stripped, _CREATE_INDEX_KW)
    if rewritten:
        return rewritten, True

    m = _ALTER_ADD_KW.match(stripped)
    if m and not _ALTER_ADD_SUBCLAUSE_RE.match(stripped[m.end() :]):
        rewritten = _insert_if_not_exists_after(stripped, _ALTER_ADD_KW)
        if rewritten:
            return rewritten, True

    if _INSERT_INTO_RE.match(stripped) and "ON CONFLICT" not in stripped.upper():
        return stripped + " ON CONFLICT DO NOTHING", True

    rewritten = _insert_if_exists_after(stripped, _DROP_TABLE_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_exists_after(stripped, _DROP_SEQ_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_exists_after(stripped, _DROP_INDEX_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_exists_after(stripped, _DROP_VIEW_KW)
    if rewritten:
        return rewritten, True

    rewritten = _insert_if_exists_after(stripped, _ALTER_DROP_COLUMN_KW)
    if rewritten:
        return rewritten, True

    return stripped, False


def make_idempotent(body):
    """Best-effort rewrite of a changeset's SQL (single- or multi-statement)
    to be safely re-runnable: split into top-level statements (dollar-quote
    aware, so DO $$ ... $$ blocks stay intact) and apply the single-
    statement guards to each independently. Returns (rewritten body, whether
    ANY statement was guarded)."""
    stripped = body.strip()
    if not stripped:
        return body, False

    statements = split_top_level_statements(stripped)
    if not statements:
        return body, False

    rewritten = []
    any_applied = False
    for stmt in statements:
        new_stmt, applied = _guard_single_statement(stmt)
        rewritten.append(new_stmt)
        any_applied = any_applied or applied

    if not any_applied:
        return body, False
    return "\n".join(f"{s};" for s in rewritten), True


def split_leading_comments(body):
    """Return (comment_block, rest) — comment_block is the leading run of
    `-- ...` / blank lines (the Liquibase <comment> text), rest is
    everything after, both preserving original text exactly."""
    lines = body.splitlines()
    i = 0
    while i < len(lines) and (lines[i].strip() == "" or lines[i].lstrip().startswith("--")):
        i += 1
    return "\n".join(lines[:i]), "\n".join(lines[i:]).strip()


def slugify(text):
    text = re.sub(r"[^a-zA-Z0-9]+", "_", text).strip("_").lower()
    return text[:60].strip("_")


def infer_down(up_body):
    """Best-effort single-statement rollback. Returns SQL string or None."""
    body = up_body.strip()
    if not body:
        return None
    if _MULTI_STATEMENT_GUARD.search(body):
        return None  # more than one statement -> don't guess
    for pattern, template in DOWN_PATTERNS:
        m = pattern.match(body)
        if m:
            return template.format(*m.groups())
    return None


def strip_leading_comments(body):
    """Return body with leading `-- ...` comment-only lines removed (those are
    the Liquibase <comment> text, not SQL) so pattern matching sees real SQL."""
    lines = body.splitlines()
    i = 0
    while i < len(lines) and (lines[i].strip() == "" or lines[i].lstrip().startswith("--")):
        i += 1
    return "\n".join(lines[i:]).strip()


def fix_known_placeholders(sql, warnings, changeset_ref):
    """Liquibase changelog property placeholders that offline mode leaves
    unresolved. Confirmed via grep: exactly one occurrence in this changelog
    (validation_site_information.xml::1::caleb), '${now}' meant to be a real
    timestamp function call."""
    if "${now}" in sql:
        sql = sql.replace("'${now}'", "now()")
        warnings.append(f"{changeset_ref}: replaced unresolved '${{now}}' with now()")
    return sql


def split(full_sql_path, out_dir):
    text = Path(full_sql_path).read_text(encoding="utf-8")
    lines = text.splitlines()

    # locate marker line indices
    marker_idxs = [i for i, l in enumerate(lines) if MARKER_RE.match(l)]
    if not marker_idxs:
        raise SystemExit("no '-- Changeset ...' markers found — wrong input file?")

    out_dir = Path(out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    warnings = []
    no_txn_count = 0
    auto_down_count = 0
    todo_down_count = 0
    guarded_count = 0
    unguarded_count = 0
    manifest = []

    for n, idx in enumerate(marker_idxs, start=1):
        m = MARKER_RE.match(lines[idx])
        src_path, cs_id, author = m.group("path"), m.group("id"), m.group("author")
        end = marker_idxs[n] if n < len(marker_idxs) else len(lines)
        raw_body = "\n".join(lines[idx + 1 : end]).strip("\n")

        changeset_ref = f"{src_path}::{cs_id}::{author}"
        raw_body = fix_known_placeholders(raw_body, warnings, changeset_ref)

        up_only = strip_leading_comments(raw_body)
        down_sql = infer_down(up_only)

        no_txn = bool(NO_TXN_RE.search(raw_body))
        if no_txn:
            no_txn_count += 1

        comment_block, sql_part = split_leading_comments(raw_body)
        guarded_sql, was_guarded = make_idempotent(sql_part)
        up_text = (comment_block + "\n" + guarded_sql).strip() if comment_block else guarded_sql
        if was_guarded:
            guarded_count += 1
        else:
            unguarded_count += 1

        file_stem = Path(src_path).stem
        slug = slugify(f"{file_stem}_{cs_id}")
        fname = f"{n:04d}_{slug}.sql"

        # Every changeset body is wrapped in StatementBegin/StatementEnd so
        # goose treats it as ONE opaque statement sent verbatim to the
        # driver, instead of goose's own naive semicolon-splitter — which
        # breaks on any internal ';' (DO $$ ... $$ blocks, multi-statement
        # bodies) and produces "unterminated dollar-quoted string" errors.
        # Confirmed empirically: 0001_convert_id_types_1.sql (a DO $$ block)
        # failed exactly this way before this wrapping was added. Postgres
        # itself is fine executing multiple ';'-separated statements in one
        # Exec() call (simple query protocol) — only goose's own splitting
        # needs suppressing.
        parts = []
        parts.append(f"-- source: liquibase {changeset_ref}\n")
        if no_txn:
            parts.append("-- +goose NO TRANSACTION\n")
        parts.append("-- +goose Up\n")
        parts.append("-- +goose StatementBegin\n")
        parts.append(up_text + "\n")
        parts.append("-- +goose StatementEnd\n")
        parts.append("\n-- +goose Down\n")
        if down_sql:
            parts.append("-- +goose StatementBegin\n")
            parts.append(down_sql + "\n")
            parts.append("-- +goose StatementEnd\n")
            auto_down_count += 1
        else:
            parts.append(
                "-- TODO: no safe auto-generated rollback for this changeset.\n"
                f"-- Liquibase source: {changeset_ref}\n"
                "-- Hand-write if this migration must be reversible; see\n"
                "-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).\n"
            )
            todo_down_count += 1

        (out_dir / fname).write_text("".join(parts), encoding="utf-8", newline="\n")
        manifest.append((fname, changeset_ref, no_txn, down_sql is not None, was_guarded))

    print(f"wrote {len(marker_idxs)} migration files to {out_dir}")
    print(f"  NO TRANSACTION: {no_txn_count}")
    print(f"  auto-generated Down: {auto_down_count}")
    print(f"  TODO Down (needs hand authorship): {todo_down_count}")
    print(f"  idempotency-guarded (IF NOT EXISTS / ON CONFLICT DO NOTHING): {guarded_count}")
    print(f"  NOT idempotency-guarded (multi-statement, needs hand review): {unguarded_count}")
    if warnings:
        print(f"  warnings ({len(warnings)}):")
        for w in warnings:
            print(f"    - {w}")

    manifest_path = out_dir.parent / "MIGRATION_MANIFEST.tsv"
    with open(manifest_path, "w", encoding="utf-8") as f:
        f.write("file\tliquibase_source\tno_transaction\thas_down\tidempotency_guarded\n")
        for fname, ref, no_txn, has_down, guarded in manifest:
            f.write(f"{fname}\t{ref}\t{no_txn}\t{has_down}\t{guarded}\n")
    print(f"manifest: {manifest_path}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(1)
    split(sys.argv[1], sys.argv[2])
