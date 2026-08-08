#!/usr/bin/env python3
"""Load db/dbInit/OpenELIS-Global.sql (a pg_dump schema+seed-data dump) into
a Postgres database — the same "genesis" baseline the Docker database image
and the install scripts use. This is NOT part of the Liquibase changelog
tree; Liquibase's 2.0.x.x-onward changesets (and the goose migrations split
from them, see split_liquibase_sql.py) all assume this baseline already
exists — exactly as they do in production today.

Needed because the dump uses `COPY ... FROM stdin; <data> \\.` blocks for
bulk data (this is how pg_dump emits seed rows), which is a psql/libpq
streaming protocol construct, not something a plain string Exec() can run.
This script splits the file into plain-SQL chunks and COPY chunks and drives
each through the right psycopg2 API (execute() vs copy_expert()).

Usage:
    pip install psycopg2-binary
    python3 load_baseline_dump.py <dump.sql> "<dsn>"

<dsn> example: postgres://postgres:admin@localhost:15432/clinlims_goose_verify
"""

import io
import re
import sys

import psycopg2

COPY_HEADER_RE = re.compile(r"^COPY\s+(?P<target>.*?)\s+FROM\s+stdin;\s*$", re.IGNORECASE)

# Some tables in this dump have their COPY block deliberately commented out
# with /* ... */ (data stripped from the "genesis" baseline, e.g. for size).
# The fake header inside a comment ("/*COPY foo FROM stdin;") must NOT match
# COPY_HEADER_RE (it doesn't — the regex is anchored to a line that STARTS
# with COPY) but a whole stretch of the dump can end up being nothing but
# '--' line comments and one or more of these /* */ blocks, with zero real
# statements. Sending that to psycopg2 raises "can't execute an empty query"
# — so block comments need stripping too when checking for real content.
BLOCK_COMMENT_RE = re.compile(r"/\*.*?\*/", re.DOTALL)


def has_executable_sql(body):
    no_block_comments = BLOCK_COMMENT_RE.sub("", body)
    remaining = "\n".join(
        l for l in no_block_comments.splitlines() if l.strip() and not l.strip().startswith("--")
    ).strip()
    return bool(remaining)


def split_chunks(text):
    lines = text.split("\n")
    chunks = []
    buf = []
    i, n = 0, len(lines)
    while i < n:
        line = lines[i]
        m = COPY_HEADER_RE.match(line.strip())
        if m:
            if buf:
                chunks.append(("sql", "\n".join(buf)))
                buf = []
            i += 1
            data_lines = []
            while i < n and lines[i] != "\\.":
                data_lines.append(lines[i])
                i += 1
            i += 1  # skip the '\.' terminator line
            chunks.append(("copy", m.group("target"), "\n".join(data_lines)))
            continue
        buf.append(line)
        i += 1
    if buf:
        chunks.append(("sql", "\n".join(buf)))
    return chunks


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(1)
    dump_path, dsn = sys.argv[1], sys.argv[2]

    text = open(dump_path, encoding="utf-8").read()
    chunks = split_chunks(text)
    sql_chunks = sum(1 for c in chunks if c[0] == "sql")
    copy_chunks = sum(1 for c in chunks if c[0] == "copy")
    print(f"parsed {len(chunks)} chunks ({sql_chunks} sql, {copy_chunks} copy)")

    conn = psycopg2.connect(dsn)
    conn.autocommit = False
    cur = conn.cursor()
    try:
        for chunk_idx, (kind, *rest) in enumerate(chunks):
            if kind == "sql":
                body = rest[0].strip()
                if not has_executable_sql(body):
                    continue
                cur.execute(body)
            else:
                target, data = rest
                cur.copy_expert(f"COPY {target} FROM STDIN", io.StringIO(data + "\n" if data else ""))
        conn.commit()
    except Exception:
        conn.rollback()
        dump_path_out = "E:/Temp/claude/E--workspace-migration-OpenELIS-Global-2-go/2faa2b89-8986-4276-a8ff-affeaca0e7eb/scratchpad/failed_chunk.txt"
        with open(dump_path_out, "w", encoding="utf-8") as f:
            f.write(f"FAILED at chunk {chunk_idx} ({chunks[chunk_idx][0]})\n")
            f.write(repr(chunks[chunk_idx]))
        print(f"FAILED at chunk {chunk_idx} ({chunks[chunk_idx][0]}) -- details written to {dump_path_out}")
        raise
    print("baseline loaded OK")


if __name__ == "__main__":
    main()
