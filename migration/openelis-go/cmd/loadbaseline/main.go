// Command loadbaseline loads db/dbInit/OpenELIS-Global.sql (a pg_dump
// schema+seed-data dump) into a Postgres database — the same "genesis"
// baseline the Docker database image and the install scripts use. This is
// NOT part of the Liquibase changelog tree; Liquibase's 2.0.x.x-onward
// changesets (and the goose migrations split from them, see
// cmd/splitliquibase) all assume this baseline already exists — exactly as
// they do in production today.
//
// Needed because the dump uses `COPY ... FROM stdin; <data> \.` blocks for
// bulk data (this is how pg_dump emits seed rows), which is a psql/libpq
// streaming protocol construct, not something a plain string Exec() can
// run. This splits the file into plain-SQL chunks and COPY chunks, and
// drives the COPY chunks through lib/pq's CopyIn protocol support
// (pq.CopyInSchema) instead — which takes already-parsed field values, not
// raw COPY-format text, so each data row is unescaped per Postgres's COPY
// TEXT format before being handed to it.
//
// Usage:
//
//	go run ./cmd/loadbaseline <dump.sql> "<dsn>"
//
// <dsn> example: postgres://postgres:admin@localhost:15432/clinlims_goose_verify
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

var (
	copyHeaderRE = regexp.MustCompile(`(?i)^COPY\s+(.*?)\s+FROM\s+stdin;\s*$`)

	// Some tables in this dump have their COPY block deliberately commented
	// out with /* ... */ (data stripped from the "genesis" baseline, e.g.
	// for size). The fake header inside a comment ("/*COPY foo FROM
	// stdin;") must NOT match copyHeaderRE (it doesn't — the regex is
	// anchored to a line that STARTS with COPY) but a whole stretch of the
	// dump can end up being nothing but '--' line comments and one or more
	// of these /* */ blocks, with zero real statements. Sending that to
	// Postgres raises a syntax error on an empty query — so block comments
	// need stripping too when checking for real content.
	blockCommentRE = regexp.MustCompile(`(?s)/\*.*?\*/`)

	// clinlims.organization (id, name, ...)  ->  schema=clinlims table=organization cols=[id name ...]
	copyTargetRE = regexp.MustCompile(`^(?:([\w]+)\.)?([\w]+)\s*\(([^)]*)\)$`)
)

type chunk struct {
	kind    string // "sql" or "copy"
	sqlBody string
	target  string // copy only: raw "schema.table (col1, col2, ...)"
	data    string // copy only: raw tab-separated data lines joined by \n
}

func hasExecutableSQL(body string) bool {
	noBlockComments := blockCommentRE.ReplaceAllString(body, "")
	var kept []string
	for _, l := range strings.Split(noBlockComments, "\n") {
		t := strings.TrimSpace(l)
		if t != "" && !strings.HasPrefix(t, "--") {
			kept = append(kept, l)
		}
	}
	return strings.TrimSpace(strings.Join(kept, "\n")) != ""
}

func splitChunks(text string) []chunk {
	lines := strings.Split(text, "\n")
	var chunks []chunk
	var buf []string
	i, n := 0, len(lines)
	for i < n {
		line := lines[i]
		if m := copyHeaderRE.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			if len(buf) > 0 {
				chunks = append(chunks, chunk{kind: "sql", sqlBody: strings.Join(buf, "\n")})
				buf = nil
			}
			i++
			var dataLines []string
			for i < n && lines[i] != `\.` {
				dataLines = append(dataLines, lines[i])
				i++
			}
			i++ // skip the '\.' terminator line
			chunks = append(chunks, chunk{kind: "copy", target: m[1], data: strings.Join(dataLines, "\n")})
			continue
		}
		buf = append(buf, line)
		i++
	}
	if len(buf) > 0 {
		chunks = append(chunks, chunk{kind: "sql", sqlBody: strings.Join(buf, "\n")})
	}
	return chunks
}

// unescapeCopyField reverses Postgres's COPY TEXT format escaping for one
// field. Returns (value, isNull).
func unescapeCopyField(s string) (string, bool) {
	if s == `\N` {
		return "", true
	}
	var b strings.Builder
	i, n := 0, len(s)
	for i < n {
		c := s[i]
		if c == '\\' && i+1 < n {
			switch s[i+1] {
			case '\\':
				b.WriteByte('\\')
			case 't':
				b.WriteByte('\t')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'v':
				b.WriteByte('\v')
			default:
				b.WriteByte(s[i+1])
			}
			i += 2
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String(), false
}

// parseCopyTarget splits "clinlims.organization (id, name, ...)" into
// schema, table, and column names.
func parseCopyTarget(target string) (schema, table string, cols []string, err error) {
	m := copyTargetRE.FindStringSubmatch(strings.TrimSpace(target))
	if m == nil {
		return "", "", nil, fmt.Errorf("could not parse COPY target %q", target)
	}
	schema, table = m[1], m[2]
	for _, c := range strings.Split(m[3], ",") {
		cols = append(cols, strings.Trim(strings.TrimSpace(c), `"`))
	}
	return schema, table, cols, nil
}

func loadCopyChunk(tx *sql.Tx, c chunk) error {
	schema, table, cols, err := parseCopyTarget(c.target)
	if err != nil {
		return err
	}
	var copySQL string
	if schema != "" {
		copySQL = pq.CopyInSchema(schema, table, cols...)
	} else {
		copySQL = pq.CopyIn(table, cols...)
	}
	stmt, err := tx.Prepare(copySQL)
	if err != nil {
		return fmt.Errorf("prepare copy for %s: %w", c.target, err)
	}
	if c.data != "" {
		for _, line := range strings.Split(c.data, "\n") {
			fields := strings.Split(line, "\t")
			if len(fields) != len(cols) {
				return fmt.Errorf("copy %s: row has %d fields, expected %d: %q", c.target, len(fields), len(cols), line)
			}
			values := make([]interface{}, len(fields))
			for i, f := range fields {
				v, isNull := unescapeCopyField(f)
				if isNull {
					values[i] = nil
				} else {
					values[i] = v
				}
			}
			if _, err := stmt.Exec(values...); err != nil {
				stmt.Close()
				return fmt.Errorf("copy %s row: %w", c.target, err)
			}
		}
	}
	if _, err := stmt.Exec(); err != nil {
		stmt.Close()
		return fmt.Errorf("copy %s finish: %w", c.target, err)
	}
	return stmt.Close()
}

func run(dumpPath, dsn string) error {
	data, err := os.ReadFile(dumpPath)
	if err != nil {
		return err
	}
	chunks := splitChunks(string(data))
	sqlCount, copyCount := 0, 0
	for _, c := range chunks {
		if c.kind == "sql" {
			sqlCount++
		} else {
			copyCount++
		}
	}
	fmt.Printf("parsed %d chunks (%d sql, %d copy)\n", len(chunks), sqlCount, copyCount)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for idx, c := range chunks {
		var execErr error
		if c.kind == "sql" {
			body := strings.TrimSpace(c.sqlBody)
			if !hasExecutableSQL(body) {
				continue
			}
			_, execErr = tx.Exec(body)
		} else {
			execErr = loadCopyChunk(tx, c)
		}
		if execErr != nil {
			tx.Rollback()
			// Written beside the input dump — deterministic and portable,
			// not a hardcoded scratch path (an earlier draft of the
			// Python version of this tool hardcoded a session-specific
			// path here; opening it raised FileNotFoundError on any other
			// checkout, masking the real Postgres failure being
			// diagnosed).
			diagPath := filepath.Join(filepath.Dir(dumpPath), strings.TrimSuffix(filepath.Base(dumpPath), filepath.Ext(dumpPath))+".failed_chunk.txt")
			diagContent := fmt.Sprintf("FAILED at chunk %d (%s)\nerror: %v\n\n%+v\n", idx, c.kind, execErr, c)
			_ = os.WriteFile(diagPath, []byte(diagContent), 0o644)
			fmt.Printf("FAILED at chunk %d (%s) -- details written to %s\n", idx, c.kind, diagPath)
			return execErr
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("baseline loaded OK")
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: loadbaseline <dump.sql> <dsn>")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
