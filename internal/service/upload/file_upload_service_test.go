package upload

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
)

func setupFileUploadTest(t *testing.T) (*FileUploadService, *gorm.DB, *store.FsStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.FileRecord{}, &model.AuditLog{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	fsStore, err := store.NewFsStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create fs store: %v", err)
	}

	validator := NewFileValidator(
		[]string{".zip", ".tar.gz", ".tgz", ".7z", ".rar"},
		[]string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"},
	)

	fileRecordRepo := repo.NewFileRecordRepo(db)
	auditRepo := repo.NewAuditRepo(db)
	auditSvc := audit.NewService(auditRepo)

	svc := NewFileUploadService(validator, fsStore, fileRecordRepo, auditSvc, db, 20*1024*1024)
	return svc, db, fsStore
}

func TestFileUploadService_Upload_Success(t *testing.T) {
	svc, db, _ := setupFileUploadTest(t)

	data := []byte("file content data")
	in := FileUploadInput{
		Purpose:  string(model.FilePurposeAvatar),
		Filename: "avatar.jpg",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.FileId == "" {
		t.Error("FileId should not be empty")
	}
	if result.URL == "" {
		t.Error("URL should not be empty")
	}
	if !strings.HasPrefix(result.URL, "/files/avatar/") {
		t.Errorf("URL = %q, should start with /files/avatar/", result.URL)
	}
	if result.Checksum == "" {
		t.Error("Checksum should not be empty")
	}
	if len(result.Checksum) != 64 {
		t.Errorf("Checksum length = %d, want 64", len(result.Checksum))
	}
	if result.Size != int64(len(data)) {
		t.Errorf("Size = %d, want %d", result.Size, len(data))
	}
	if result.Purpose != string(model.FilePurposeAvatar) {
		t.Errorf("Purpose = %q, want %q", result.Purpose, model.FilePurposeAvatar)
	}
	if result.UploadedAt == "" {
		t.Error("UploadedAt should not be empty")
	}

	var record model.FileRecord
	if err := db.Where("id = ?", result.FileId).First(&record).Error; err != nil {
		t.Fatalf("query file record failed: %v", err)
	}
	if record.URL != result.URL {
		t.Errorf("record URL = %q, want %q", record.URL, result.URL)
	}
	if record.UploaderIdentity != "uploader1" {
		t.Errorf("UploaderIdentity = %q, want %q", record.UploaderIdentity, "uploader1")
	}

	var logs []model.AuditLog
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("query audit logs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs count = %d, want 1", len(logs))
	}
	if logs[0].Action != model.AuditActionFileUploaded {
		t.Errorf("audit Action = %q, want %q", logs[0].Action, model.AuditActionFileUploaded)
	}
}

func TestFileUploadService_Upload_URLContainsUUID(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	data := []byte("uuid test")
	in := FileUploadInput{
		Purpose:  string(model.FilePurposeScreenshot),
		Filename: "screen.png",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	parts := strings.Split(result.URL, "/")
	if len(parts) < 5 {
		t.Fatalf("URL parts = %d, want at least 5", len(parts))
	}

	uuidPart := parts[3]
	if len(uuidPart) != 36 {
		t.Errorf("UUID part length = %d, want 36 (standard UUID format)", len(uuidPart))
	}

	if uuidPart == "screen.png" || uuidPart == "screenshot" {
		t.Error("UUID part should be a random UUID, not the filename or purpose")
	}
}

func TestFileUploadService_Upload_AllPurposes(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	purposes := []string{
		string(model.FilePurposeAvatar),
		string(model.FilePurposeAttachment),
		string(model.FilePurposeScreenshot),
		string(model.FilePurposeDocument),
		string(model.FilePurposeOther),
	}

	for _, purpose := range purposes {
		t.Run(purpose, func(t *testing.T) {
			data := []byte("test data for " + purpose)
			in := FileUploadInput{
				Purpose:  purpose,
				Filename: "test.jpg",
				FileSize: int64(len(data)),
			}
			result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if err != nil {
				t.Errorf("Upload with purpose=%q failed: %v", purpose, err)
			}
			if result != nil && !strings.HasPrefix(result.URL, "/files/"+purpose+"/") {
				t.Errorf("URL = %q, should start with /files/%s/", result.URL, purpose)
			}
		})
	}
}

func TestFileUploadService_Upload_InvalidPurpose(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	tests := []struct {
		name    string
		purpose string
	}{
		{"unknown purpose", "unknown"},
		{"empty purpose", ""},
		{"hack purpose", "hack"},
		{"software purpose (not for files)", "software"},
		{"plugin purpose (not for files)", "plugin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("test")
			in := FileUploadInput{
				Purpose:  tt.purpose,
				Filename: "test.jpg",
				FileSize: int64(len(data)),
			}
			_, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if !errors.Is(err, ErrInvalidPurpose) {
				t.Errorf("Upload with purpose=%q error = %v, want ErrInvalidPurpose", tt.purpose, err)
			}
		})
	}
}

func TestFileUploadService_Upload_InvalidFileType(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	tests := []struct {
		name     string
		filename string
	}{
		{"exe file", "malware.exe"},
		{"sh file", "script.sh"},
		{"zip file (not for general)", "archive.zip"},
		{"no extension", "file"},
		{"js file", "script.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("test")
			in := FileUploadInput{
				Purpose:  string(model.FilePurposeAvatar),
				Filename: tt.filename,
				FileSize: int64(len(data)),
			}
			_, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if !errors.Is(err, ErrInvalidFileType) {
				t.Errorf("Upload with filename=%q error = %v, want ErrInvalidFileType", tt.filename, err)
			}
		})
	}
}

func TestFileUploadService_Upload_FileTooLarge(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	in := FileUploadInput{
		Purpose:  string(model.FilePurposeAvatar),
		Filename: "big.jpg",
		FileSize: 25 * 1024 * 1024,
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("Upload error = %v, want ErrFileTooLarge", err)
	}
}

func TestFileUploadService_Upload_FilenameSanitization(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	tests := []struct {
		name      string
		filename  string
		wantInURL bool
	}{
		{"path traversal", "../../etc/passwd.jpg", true},
		{"windows path", "..\\..\\windows\\system32.png", true},
		{"with directory", "dir/photo.jpg", true},
		{"simple filename", "photo.jpg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("sanitized test")
			in := FileUploadInput{
				Purpose:  string(model.FilePurposeAvatar),
				Filename: tt.filename,
				FileSize: int64(len(data)),
			}
			result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if err != nil {
				t.Fatalf("Upload failed: %v", err)
			}

			if strings.Contains(result.URL, "..") {
				t.Errorf("URL contains '..': %q", result.URL)
			}
			if strings.Contains(result.URL, "\\") {
				t.Errorf("URL contains backslash: %q", result.URL)
			}
		})
	}
}

func TestFileUploadService_Upload_ChecksumCorrect(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	data := []byte("file checksum verification test")
	in := FileUploadInput{
		Purpose:  string(model.FilePurposeDocument),
		Filename: "doc.pdf",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	expected := computeExpectedChecksum(t, data)
	if result.Checksum != expected {
		t.Errorf("Checksum = %q, want %q", result.Checksum, expected)
	}
}

func TestFileUploadService_Upload_MultipleFilesDifferentURLs(t *testing.T) {
	svc, _, _ := setupFileUploadTest(t)

	data := []byte("same content")
	in := FileUploadInput{
		Purpose:  string(model.FilePurposeAvatar),
		Filename: "same.jpg",
		FileSize: int64(len(data)),
	}

	result1, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	result2, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if err != nil {
		t.Fatalf("second Upload failed: %v", err)
	}

	if result1.URL == result2.URL {
		t.Error("two uploads with same filename should have different URLs (different UUID paths)")
	}
	if result1.FileId == result2.FileId {
		t.Error("two uploads should have different FileIds")
	}
}
