package jobs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeJobExecuteSetsCompleted(t *testing.T) {
	job := &ComputeJob{JobID: "j1", Image: "img:test"}
	if err := job.Execute([]byte("payload")); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := job.GetStatus(); got != "completed" {
		t.Fatalf("expected completed state, got %q", got)
	}
}

func TestStorageJobExecuteWithoutS3ClientStillCompletes(t *testing.T) {
	oldClient := s3Client
	s3Client = nil
	t.Cleanup(func() { s3Client = oldClient })

	job := &StorageJob{JobID: "job-a", ChunkID: "chunk-a"}
	if err := job.Execute([]byte("abc123")); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := job.GetStatus(); got != "completed" {
		t.Fatalf("expected completed state, got %q", got)
	}
}

func TestWriteFilePersistsBytes(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "chunk.bin")
	data := []byte("encrypted-bytes")

	if err := writeFile(target, data); err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected file content: got %q want %q", string(got), string(data))
	}
}

func TestWriteFileInvalidPathReturnsError(t *testing.T) {
	invalid := filepath.Join(t.TempDir(), "missing", "chunk.bin")
	if err := writeFile(invalid, []byte("data")); err == nil {
		t.Fatalf("expected error for invalid path")
	}
}
