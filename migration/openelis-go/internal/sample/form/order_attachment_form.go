package form

// OrderAttachmentDTO mirrors OrderAttachmentRestController.toDto — a Map.of
// with exactly these five keys.
//
// Map.of REJECTS null values (it throws NullPointerException), which is why
// Java coalesces every nullable field at the call site: fileType null -> "",
// fileSizeBytes null -> 0L, uploadedAt null -> "". So none of these keys is
// ever absent and none is ever null, and none takes omitempty here.
//
// uploadedAt is Timestamp.toString(), i.e. "2026-08-28 09:15:22.123" — a
// space-separated local rendering, not ISO-8601 and not epoch millis. A port
// that formats it as RFC3339 diverges on every row.
type OrderAttachmentDTO struct {
	ID            int64  `json:"id"`
	FileName      string `json:"fileName"`
	FileType      string `json:"fileType"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	UploadedAt    string `json:"uploadedAt"`
}

// ErrorDTO is Java's Map.of("error", "...") body, used by the attachment
// endpoints' 404.
type ErrorDTO struct {
	Error string `json:"error"`
}
