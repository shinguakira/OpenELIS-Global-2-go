// Command splitliquibase splits a Liquibase `updateSql`/offline-mode SQL
// dump into numbered goose migration files.
//
// Input is the file produced by (see migration/liquibase-to-goose-plan.md § 2):
//
//	liquibase --classpath=src/main/resources \
//	          --changelog-file=liquibase/base-changelog.xml \
//	          --output-file=<full.sql> \
//	          updateSQL --url=offline:postgresql
//
// Each changeset in that file is preceded by a marker line of the exact form:
//
//	-- Changeset <changelog/relative/path.xml>::<id>::<author>
//
// This splits on that marker, and for every changeset emits one file
//
//	db/migrations/<seq>_<slug>.sql
//
// in goose format (`-- +goose Up` / `-- +goose Down`).
//
// Down-block generation is a best-effort heuristic (see inferDown) for the
// small set of single-statement, unambiguous patterns (CREATE TABLE, CREATE
// SEQUENCE, CREATE INDEX, single-column ADD COLUMN, CREATE VIEW). Everything
// else — multi-statement bodies, DML, ALTER COLUMN TYPE, DROP, DO blocks,
// added constraints — gets an honest `-- TODO` placeholder instead of a
// guessed rollback. See migration/liquibase-to-goose-plan.md § 7 "Risk
// items" for why: guessing wrong here is worse than admitting the gap.
//
// Usage:
//
//	go run ./cmd/splitliquibase <full.sql> <out_dir>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	markerRE = regexp.MustCompile(`^-- Changeset (\S+)::([^:]+)::(.+)$`)

	// Statements needing NO TRANSACTION in Postgres (can't run inside a tx block).
	noTxnRE = regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY|DROP\s+INDEX\s+CONCURRENTLY|ALTER\s+TYPE\s+\S+\s+ADD\s+VALUE`)

	// A ';' followed by more non-whitespace content -> more than one statement.
	multiStatementGuard = regexp.MustCompile(`;\s*\S`)

	// --- best-effort Down inference (single-statement, unambiguous shapes only) ---
	createTableDownRE = regexp.MustCompile(`(?is)^CREATE TABLE\s+(?:IF NOT EXISTS\s+)?([\w.]+)\s*\(.*\)\s*;?\s*$`)
	createViewDownRE  = regexp.MustCompile(`(?is)^CREATE\s+(?:OR REPLACE\s+)?VIEW\s+([\w.]+)\s+AS\b.*$`)
	createSeqDownRE   = regexp.MustCompile(`(?is)^CREATE\s+SEQUENCE\s+(?:IF NOT EXISTS\s+)?([\w.]+).*$`)
	createIndexDownRE = regexp.MustCompile(`(?is)^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:CONCURRENTLY\s+)?(?:IF NOT EXISTS\s+)?([\w."]+)\s+ON\b.*$`)

	// --- idempotency guards ---
	// Confirmed empirically (running the generated migrations against a
	// clean DB, see migration/liquibase-to-goose-plan.md): plain extracted
	// SQL hits real "duplicate key value violates unique constraint" errors
	// on ordinary INSERTs. Root cause: 218 of the source XML files use
	// <preConditions onFail="MARK_RAN">, which Liquibase evaluates against
	// live DB state at apply time (skip silently if the precondition
	// fails) — offline-mode extraction has no DB to check, so it can't
	// reproduce that skip and just always emits the statement. These
	// rewrites make the safely-rewritable statements idempotent (CREATE
	// TABLE/SEQUENCE/INDEX -> IF NOT EXISTS, ADD COLUMN -> IF NOT EXISTS,
	// plain INSERT -> ON CONFLICT DO NOTHING), the same practical effect
	// MARK_RAN had — applied per-statement within a changeset (see
	// splitTopLevelStatements), not just to whole-body single-statement
	// changesets. Statement shapes outside the pattern list (UPDATE,
	// DELETE, ALTER COLUMN TYPE, DROP, DO blocks, added constraints) are
	// left untouched — the manifest records which changesets got zero
	// guards so those are an itemized follow-up, not a silent gap.
	//
	// Go's regexp (RE2) has no lookahead at all, which conveniently rules
	// out the exact bug this hit in an earlier (Python) draft: a combined
	// `\s+(?!IF NOT EXISTS\b)` pattern let the engine backtrack \s+ down to
	// consuming only one of two spaces in SQL that already had irregular
	// spacing ("CREATE SEQUENCE  IF NOT EXISTS ..."), so the lookahead
	// technically found a non-match and wrongly reported "not yet guarded"
	// -> "CREATE SEQUENCE IF NOT EXISTS  IF NOT EXISTS ...". Fixed (and
	// here, structurally required) by never combining them: match only the
	// keyword prefix, then check "already guarded" as an independent regex
	// against whatever follows, tolerant of any amount of whitespace.
	createTableKW        = regexp.MustCompile(`(?i)^CREATE TABLE\s+`)
	createSeqKW          = regexp.MustCompile(`(?i)^CREATE SEQUENCE\s+`)
	createIndexKW        = regexp.MustCompile(`(?i)^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:CONCURRENTLY\s+)?`)
	alreadyIfNotExistsRE = regexp.MustCompile(`(?i)^\s*IF\s+NOT\s+EXISTS\b`)

	// DROP-side mirror: confirmed empirically (full-scan diagnostic pass)
	// that a from-scratch apply can hit a changeset DROPping something an
	// earlier changeset never created in THIS sequence (e.g. system_config,
	// a column already gone) — same underlying onFail=MARK_RAN gap,
	// opposite direction.
	dropTableKW       = regexp.MustCompile(`(?i)^DROP TABLE\s+`)
	dropSeqKW         = regexp.MustCompile(`(?i)^DROP SEQUENCE\s+`)
	dropIndexKW       = regexp.MustCompile(`(?i)^DROP\s+INDEX\s+(?:CONCURRENTLY\s+)?`)
	dropViewKW        = regexp.MustCompile(`(?i)^DROP VIEW\s+`)
	alterDropColumnKW = regexp.MustCompile(`(?i)^ALTER TABLE\s+[\w.]+\s+DROP\s+COLUMN\s+`)
	alreadyIfExistsRE = regexp.MustCompile(`(?i)^\s*IF\s+EXISTS\b`)

	// "ADD" on ALTER TABLE is ambiguous — ADD COLUMN vs ADD CONSTRAINT/
	// PRIMARY KEY/UNIQUE/FOREIGN KEY/CHECK/EXCLUDE, and only ADD [COLUMN]
	// accepts IF NOT EXISTS in Postgres, and only ADD [COLUMN] safely
	// inverts to DROP COLUMN for down-inference. Confirmed via TWO real
	// bugs this exclusion fixes: (1) a syntax error empirically hitting
	// "ADD IF NOT EXISTS CONSTRAINT ... UNIQUE" before the guard-side
	// exclusion was added, and (2) a Codex review catching migration 0856
	// emitting "ALTER TABLE clinlims.analysis DROP COLUMN IF EXISTS
	// CONSTRAINT" as a bogus down-inference — CONSTRAINT was matched as if
	// it were the column name, producing a silent no-op rollback that left
	// the real FK constraint in place while goose recorded the migration
	// as reverted, so a later redo failed re-adding a constraint that was
	// never actually dropped.
	alterAddKW          = regexp.MustCompile(`(?i)^ALTER TABLE\s+[\w.]+\s+ADD\s+(?:COLUMN\s+)?`)
	alterAddSubclauseRE = regexp.MustCompile(`(?i)^\s*(?:CONSTRAINT|PRIMARY|UNIQUE|FOREIGN|CHECK|EXCLUDE)\b`)
	alterAddColNameRE   = regexp.MustCompile(`^\s*("?\w+"?)\s+\S`)
	alterTableNameRE    = regexp.MustCompile(`(?i)^ALTER TABLE\s+([\w.]+)`)

	insertIntoRE = regexp.MustCompile(`(?i)^INSERT\s+INTO\b`)

	dollarTagRE = regexp.MustCompile(`^\$[A-Za-z_]*\$`)

	slugNonAlnumRE = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// insertGuardAfter: if s starts with kw and isn't already followed by
// guardPhrase (per alreadyRE, tolerating any whitespace), returns the
// rewritten string with guardPhrase inserted; else "", false.
func insertGuardAfter(s string, kw *regexp.Regexp, alreadyRE *regexp.Regexp, guardPhrase string) (string, bool) {
	loc := kw.FindStringIndex(s)
	if loc == nil {
		return "", false
	}
	end := loc[1]
	if alreadyRE.MatchString(s[end:]) {
		return "", false
	}
	return s[:end] + guardPhrase + " " + s[end:], true
}

func insertIfNotExistsAfter(s string, kw *regexp.Regexp) (string, bool) {
	return insertGuardAfter(s, kw, alreadyIfNotExistsRE, "IF NOT EXISTS")
}

func insertIfExistsAfter(s string, kw *regexp.Regexp) (string, bool) {
	return insertGuardAfter(s, kw, alreadyIfExistsRE, "IF EXISTS")
}

// splitTopLevelStatements splits on ';' while respecting '...'/"..." string
// literals and $$.../$tag$...$tag$ dollar-quoting (so a DO $$ ... ; ... $$
// block, whose body legitimately contains ';', is never split apart).
// Returns statement strings with the separating ';' removed; a final
// statement with no trailing ';' is included as-is.
func splitTopLevelStatements(sql string) []string {
	var stmts []string
	var buf strings.Builder
	i, n := 0, len(sql)
	dollarTag := ""
	inSquote := false
	inDquote := false
	for i < n {
		c := sql[i]
		if dollarTag != "" {
			if strings.HasPrefix(sql[i:], dollarTag) {
				buf.WriteString(dollarTag)
				i += len(dollarTag)
				dollarTag = ""
			} else {
				buf.WriteByte(c)
				i++
			}
			continue
		}
		if inSquote {
			buf.WriteByte(c)
			if c == '\'' {
				if i+1 < n && sql[i+1] == '\'' {
					buf.WriteByte('\'')
					i += 2
					continue
				}
				inSquote = false
			}
			i++
			continue
		}
		if inDquote {
			buf.WriteByte(c)
			if c == '"' {
				inDquote = false
			}
			i++
			continue
		}
		switch c {
		case '\'':
			inSquote = true
			buf.WriteByte(c)
			i++
			continue
		case '"':
			inDquote = true
			buf.WriteByte(c)
			i++
			continue
		case '$':
			if loc := dollarTagRE.FindString(sql[i:]); loc != "" {
				dollarTag = loc
				buf.WriteString(dollarTag)
				i += len(dollarTag)
				continue
			}
		case ';':
			stmts = append(stmts, buf.String())
			buf.Reset()
			i++
			continue
		}
		buf.WriteByte(c)
		i++
	}
	tail := strings.TrimSpace(buf.String())
	if tail != "" {
		stmts = append(stmts, tail)
	}
	out := stmts[:0]
	for _, s := range stmts {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// guardSingleStatement applies one idempotency rewrite to a single
// already-isolated statement (no trailing ';'). Returns the possibly-
// rewritten statement and whether a guard was applied.
func guardSingleStatement(s string) (string, bool) {
	if r, ok := insertIfNotExistsAfter(s, createTableKW); ok {
		return r, true
	}
	if r, ok := insertIfNotExistsAfter(s, createSeqKW); ok {
		return r, true
	}
	if r, ok := insertIfNotExistsAfter(s, createIndexKW); ok {
		return r, true
	}
	if loc := alterAddKW.FindStringIndex(s); loc != nil && !alterAddSubclauseRE.MatchString(s[loc[1]:]) {
		if r, ok := insertIfNotExistsAfter(s, alterAddKW); ok {
			return r, true
		}
	}
	if insertIntoRE.MatchString(s) && !strings.Contains(strings.ToUpper(s), "ON CONFLICT") {
		return s + " ON CONFLICT DO NOTHING", true
	}
	if r, ok := insertIfExistsAfter(s, dropTableKW); ok {
		return r, true
	}
	if r, ok := insertIfExistsAfter(s, dropSeqKW); ok {
		return r, true
	}
	if r, ok := insertIfExistsAfter(s, dropIndexKW); ok {
		return r, true
	}
	if r, ok := insertIfExistsAfter(s, dropViewKW); ok {
		return r, true
	}
	if r, ok := insertIfExistsAfter(s, alterDropColumnKW); ok {
		return r, true
	}
	return s, false
}

// makeIdempotent rewrites a changeset's SQL (single- or multi-statement) to
// be safely re-runnable: split into top-level statements (dollar-quote
// aware) and apply the single-statement guards to each independently.
// Returns the rewritten body and whether ANY statement was guarded.
func makeIdempotent(body string) (string, bool) {
	stripped := strings.TrimSpace(body)
	if stripped == "" {
		return body, false
	}
	statements := splitTopLevelStatements(stripped)
	if len(statements) == 0 {
		return body, false
	}
	rewritten := make([]string, len(statements))
	anyApplied := false
	for i, stmt := range statements {
		r, applied := guardSingleStatement(stmt)
		rewritten[i] = r
		anyApplied = anyApplied || applied
	}
	if !anyApplied {
		return body, false
	}
	for i := range rewritten {
		rewritten[i] += ";"
	}
	return strings.Join(rewritten, "\n"), true
}

// splitLeadingComments returns (commentBlock, rest) — commentBlock is the
// leading run of `-- ...` / blank lines (the Liquibase <comment> text), rest
// is everything after, both preserving original text exactly.
func splitLeadingComments(body string) (string, string) {
	lines := strings.Split(body, "\n")
	i := 0
	for i < len(lines) && (strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(strings.TrimLeft(lines[i], " \t"), "--")) {
		i++
	}
	return strings.Join(lines[:i], "\n"), strings.TrimSpace(strings.Join(lines[i:], "\n"))
}

func slugify(text string) string {
	s := slugNonAlnumRE.ReplaceAllString(text, "_")
	s = strings.Trim(s, "_")
	s = strings.ToLower(s)
	if len(s) > 60 {
		s = s[:60]
	}
	return strings.Trim(s, "_")
}

// inferDown is a best-effort single-statement rollback. Returns "" if none
// can be safely inferred.
func inferDown(upBody string) string {
	body := strings.TrimSpace(upBody)
	if body == "" {
		return ""
	}
	if multiStatementGuard.MatchString(body) {
		return "" // more than one statement -> don't guess
	}

	// ALTER TABLE ... ADD is handled here, not via the generic pattern
	// list below — see alterAddKW's doc comment for the real bug
	// (migration 0856, Codex review) this exclusion fixes.
	if loc := alterAddKW.FindStringIndex(body); loc != nil {
		rest := body[loc[1]:]
		if alterAddSubclauseRE.MatchString(rest) {
			return "" // ADD CONSTRAINT/PRIMARY KEY/etc — no safe inverse to guess
		}
		if m := alreadyIfNotExistsRE.FindString(rest); m != "" {
			rest = rest[len(m):]
		}
		colM := alterAddColNameRE.FindStringSubmatch(rest)
		if colM == nil {
			return ""
		}
		tableM := alterTableNameRE.FindStringSubmatch(body)
		if tableM == nil {
			return ""
		}
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;", tableM[1], colM[1])
	}

	if m := createTableDownRE.FindStringSubmatch(body); m != nil {
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;", m[1])
	}
	if m := createViewDownRE.FindStringSubmatch(body); m != nil {
		return fmt.Sprintf("DROP VIEW IF EXISTS %s;", m[1])
	}
	if m := createSeqDownRE.FindStringSubmatch(body); m != nil {
		return fmt.Sprintf("DROP SEQUENCE IF EXISTS %s;", m[1])
	}
	if m := createIndexDownRE.FindStringSubmatch(body); m != nil {
		return fmt.Sprintf("DROP INDEX IF EXISTS %s;", m[1])
	}
	return ""
}

func stripLeadingComments(body string) string {
	lines := strings.Split(body, "\n")
	i := 0
	for i < len(lines) && (strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(strings.TrimLeft(lines[i], " \t"), "--")) {
		i++
	}
	return strings.TrimSpace(strings.Join(lines[i:], "\n"))
}

// fixKnownPlaceholders rewrites Liquibase changelog property placeholders
// that offline mode leaves unresolved. Confirmed via grep: exactly one
// occurrence in this changelog (validation_site_information.xml::1::caleb),
// '${now}' meant to be a real timestamp function call.
func fixKnownPlaceholders(sql string, warnings *[]string, changesetRef string) string {
	if strings.Contains(sql, "${now}") {
		sql = strings.ReplaceAll(sql, "'${now}'", "now()")
		*warnings = append(*warnings, fmt.Sprintf("%s: replaced unresolved '${now}' with now()", changesetRef))
	}
	return sql
}

type manifestRow struct {
	fname, ref              string
	noTxn, hasDown, guarded bool
}

func run(fullSQLPath, outDir string) error {
	data, err := os.ReadFile(fullSQLPath)
	if err != nil {
		return err
	}
	// Normalize line endings before splitting. The Liquibase CLI writes
	// CRLF on Windows (confirmed: `file` reports "with CRLF, LF line
	// terminators" on the real extraction output) — Go's os.ReadFile does
	// no text-mode translation, unlike Python's universal-newlines
	// Path.read_text(), so every line would otherwise carry a trailing
	// '\r' that then leaks into changeset refs, slugs, and SQL bodies.
	text := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")

	var markerIdxs []int
	for i, l := range lines {
		if markerRE.MatchString(l) {
			markerIdxs = append(markerIdxs, i)
		}
	}
	if len(markerIdxs) == 0 {
		return fmt.Errorf("no '-- Changeset ...' markers found — wrong input file?")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	var warnings []string
	noTxnCount, autoDownCount, todoDownCount, guardedCount, unguardedCount := 0, 0, 0, 0, 0
	var manifest []manifestRow

	for n, idx := range markerIdxs {
		m := markerRE.FindStringSubmatch(lines[idx])
		srcPath, csID, author := m[1], m[2], m[3]
		end := len(lines)
		if n+1 < len(markerIdxs) {
			end = markerIdxs[n+1]
		}
		rawBody := strings.Trim(strings.Join(lines[idx+1:end], "\n"), "\n")

		changesetRef := fmt.Sprintf("%s::%s::%s", srcPath, csID, author)
		rawBody = fixKnownPlaceholders(rawBody, &warnings, changesetRef)

		upOnly := stripLeadingComments(rawBody)
		downSQL := inferDown(upOnly)

		noTxn := noTxnRE.MatchString(rawBody)
		if noTxn {
			noTxnCount++
		}

		commentBlock, sqlPart := splitLeadingComments(rawBody)
		guardedSQL, wasGuarded := makeIdempotent(sqlPart)
		upText := guardedSQL
		if commentBlock != "" {
			upText = strings.TrimSpace(commentBlock + "\n" + guardedSQL)
		}
		if wasGuarded {
			guardedCount++
		} else {
			unguardedCount++
		}

		fileStem := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		slug := slugify(fileStem + "_" + csID)
		fname := fmt.Sprintf("%04d_%s.sql", n+1, slug)

		// Every changeset body is wrapped in StatementBegin/StatementEnd so
		// goose treats it as ONE opaque statement sent verbatim to the
		// driver, instead of goose's own naive semicolon-splitter — which
		// breaks on any internal ';' (DO $$ ... $$ blocks, multi-statement
		// bodies) and produces "unterminated dollar-quoted string" errors.
		// Postgres itself is fine executing multiple ';'-separated
		// statements in one Exec() call (simple query protocol) — only
		// goose's own splitting needs suppressing.
		var b strings.Builder
		fmt.Fprintf(&b, "-- source: liquibase %s\n", changesetRef)
		if noTxn {
			b.WriteString("-- +goose NO TRANSACTION\n")
		}
		b.WriteString("-- +goose Up\n")
		b.WriteString("-- +goose StatementBegin\n")
		b.WriteString(upText + "\n")
		b.WriteString("-- +goose StatementEnd\n")
		b.WriteString("\n-- +goose Down\n")
		if downSQL != "" {
			b.WriteString("-- +goose StatementBegin\n")
			b.WriteString(downSQL + "\n")
			b.WriteString("-- +goose StatementEnd\n")
			autoDownCount++
		} else {
			b.WriteString("-- TODO: no safe auto-generated rollback for this changeset.\n")
			fmt.Fprintf(&b, "-- Liquibase source: %s\n", changesetRef)
			b.WriteString("-- Hand-write if this migration must be reversible; see\n")
			b.WriteString("-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).\n")
			todoDownCount++
		}

		if err := os.WriteFile(filepath.Join(outDir, fname), []byte(b.String()), 0o644); err != nil {
			return err
		}
		manifest = append(manifest, manifestRow{fname, changesetRef, noTxn, downSQL != "", wasGuarded})
	}

	fmt.Printf("wrote %d migration files to %s\n", len(markerIdxs), outDir)
	fmt.Printf("  NO TRANSACTION: %d\n", noTxnCount)
	fmt.Printf("  auto-generated Down: %d\n", autoDownCount)
	fmt.Printf("  TODO Down (needs hand authorship): %d\n", todoDownCount)
	fmt.Printf("  idempotency-guarded (IF NOT EXISTS / ON CONFLICT DO NOTHING): %d\n", guardedCount)
	fmt.Printf("  NOT idempotency-guarded (multi-statement, needs hand review): %d\n", unguardedCount)
	if len(warnings) > 0 {
		fmt.Printf("  warnings (%d):\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("    - %s\n", w)
		}
	}

	manifestPath := filepath.Join(filepath.Dir(outDir), "MIGRATION_MANIFEST.tsv")
	var mb strings.Builder
	mb.WriteString("file\tliquibase_source\tno_transaction\thas_down\tidempotency_guarded\n")
	for _, r := range manifest {
		fmt.Fprintf(&mb, "%s\t%s\t%s\t%s\t%s\n", r.fname, r.ref, strconv.FormatBool(r.noTxn), strconv.FormatBool(r.hasDown), strconv.FormatBool(r.guarded))
	}
	if err := os.WriteFile(manifestPath, []byte(mb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("manifest: %s\n", manifestPath)
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: splitliquibase <full.sql> <out_dir>")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
