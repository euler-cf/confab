package sync

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/ConfabulousDev/confab/pkg/logger"
)

// DryRunBackend logs sync operations locally and returns mocked backend
// responses. It must only be used for providers without backend support.
type DryRunBackend struct {
	provider string
	files    map[string]FileState
}

func NewDryRunBackend(provider string) *DryRunBackend {
	return &DryRunBackend{
		provider: provider,
		files:    make(map[string]FileState),
	}
}

func (b *DryRunBackend) Init(externalID, transcriptPath string, metadata *InitMetadata) (*InitResponse, error) {
	sessionID := "dryrun-" + b.provider + "-" + shortDryRunID(externalID)
	logger.Info("Dry-run sync init: provider=%s external_id=%s transcript=%s session_id=%s metadata=%t",
		b.provider, externalID, transcriptPath, sessionID, metadata != nil)
	return &InitResponse{
		SessionID: sessionID,
		Files:     cloneFileState(b.files),
	}, nil
}

func (b *DryRunBackend) UploadChunk(sessionID, fileName, fileType string, firstLine int, lines []string, metadata *ChunkMetadata) (int, error) {
	lastLine := firstLine + len(lines) - 1
	if len(lines) == 0 {
		lastLine = firstLine - 1
	}
	b.files[fileName] = FileState{LastSyncedLine: lastLine}
	logger.Info("Dry-run chunk upload: provider=%s session_id=%s file=%s type=%s first_line=%d lines=%d last_line=%d hash=%s metadata=%t",
		b.provider, sessionID, fileName, fileType, firstLine, len(lines), lastLine, hashLines(lines), metadata != nil)
	return lastLine, nil
}

func (b *DryRunBackend) SendEvent(sessionID, eventType string, timestamp time.Time, payload json.RawMessage) error {
	logger.Info("Dry-run event send: provider=%s session_id=%s event=%s timestamp=%s payload_bytes=%d",
		b.provider, sessionID, eventType, timestamp.Format(time.RFC3339), len(payload))
	return nil
}

func (b *DryRunBackend) UpdateSessionSummary(externalID, summary string) error {
	logger.Info("Dry-run summary update: provider=%s external_id=%s summary_bytes=%d",
		b.provider, externalID, len(summary))
	return nil
}

func shortDryRunID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	if id == "" {
		return "unknown"
	}
	return id
}

func cloneFileState(in map[string]FileState) map[string]FileState {
	out := make(map[string]FileState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func hashLines(lines []string) string {
	h := fnv.New64a()
	for _, line := range lines {
		_, _ = h.Write([]byte(line))
		_, _ = h.Write([]byte{'\n'})
	}
	return fmt.Sprintf("%016x", h.Sum64())
}
