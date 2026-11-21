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

func setupSoftwareUploadTest(t *testing.T) (*SoftwareUploadService, *gorm.DB, *store.FsStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.SoftwareVersion{}, &model.AuditLog{}); err != nil {
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

	swRepo := repo.NewSoftwareVersionRepo(db)
	auditRepo := repo.NewAuditRepo(db)
	auditSvc := audit.NewService(auditRepo)

	svc := NewSoftwareUploadService(validator, fsStore, swRepo, auditSvc, db, 500*1024*1024)
	return svc, db, fsStore
}

func TestSoftwareUploadService_Upload_Success(t *testing.T) {
	svc, db, _ := setupSoftwareUploadTest(t)

	data := bytes.Repeat([]byte("b"), 2048)
	in := SoftwareUploadInput{
		SwCode:      "myapp",
		Version:     "2.0.0",
		ReleaseNote: "new release",
		SwUrl:       "https://example.com",
		Filename:    "myapp.zip",
		FileSize:    int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.VersionId == "" {
		t.Error("VersionId should not be empty")
	}
	if result.Status != string(model.SoftwareStatusDraft) {
		t.Errorf("Status = %q, want %q", result.Status, model.SoftwareStatusDraft)
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

	var sw model.SoftwareVersion
	if err := db.Where("id = ?", result.VersionId).First(&sw).Error; err != nil {
		t.Fatalf("query software version failed: %v", err)
	}
	if sw.Status != model.SoftwareStatusDraft {
		t.Errorf("sw Status = %v, want %v", sw.Status, model.SoftwareStatusDraft)
	}
	if sw.PublishedAt != nil {
		t.Error("PublishedAt should be nil for draft status")
	}
	if sw.SwCode != "myapp" {
		t.Errorf("SwCode = %q, want %q", sw.SwCode, "myapp")
	}

	var logs []model.AuditLog
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("query audit logs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs count = %d, want 1", len(logs))
	}
	if logs[0].Action != model.AuditActionSoftwareUploaded {
		t.Errorf("audit Action = %q, want %q", logs[0].Action, model.AuditActionSoftwareUploaded)
	}
}

func TestSoftwareUploadService_Upload_MetadataValidation(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	tests := []struct {
		name string
		in   SoftwareUploadInput
		err  error
	}{
		{
			name: "swCode empty",
			in: SoftwareUploadInput{
				SwCode:   "",
				Version:  "1.0.0",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "swCode with special chars",
			in: SoftwareUploadInput{
				SwCode:   "my-app!",
				Version:  "1.0.0",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "swCode with spaces",
			in: SoftwareUploadInput{
				SwCode:   "my app",
				Version:  "1.0.0",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "invalid version",
			in: SoftwareUploadInput{
				SwCode:   "myapp",
				Version:  "v1.0",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "version with v prefix",
			in: SoftwareUploadInput{
				SwCode:   "myapp",
				Version:  "v1.0.0",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Upload(context.Background(), tt.in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
			if !errors.Is(err, tt.err) {
				t.Errorf("Upload error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestSoftwareUploadService_Upload_SwCodeValid(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	validCodes := []string{"myapp", "MY_APP", "app123", "my_app_v2"}

	for _, code := range validCodes {
		t.Run(code, func(t *testing.T) {
			data := []byte("test data")
			in := SoftwareUploadInput{
				SwCode:   code,
				Version:  "1.0.0",
				Filename: "test.zip",
				FileSize: int64(len(data)),
			}
			result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if err != nil {
				t.Errorf("Upload with swCode=%q failed: %v", code, err)
			}
			if result != nil && result.Status != string(model.SoftwareStatusDraft) {
				t.Errorf("Status = %q, want %q", result.Status, model.SoftwareStatusDraft)
			}
		})
	}
}

func TestSoftwareUploadService_Upload_InvalidFileType(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	in := SoftwareUploadInput{
		SwCode:   "myapp",
		Version:  "1.0.0",
		Filename: "myapp.exe",
		FileSize: 100,
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
	if !errors.Is(err, ErrInvalidFileType) {
		t.Errorf("Upload error = %v, want ErrInvalidFileType", err)
	}
}

func TestSoftwareUploadService_Upload_FileTooLarge(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	in := SoftwareUploadInput{
		SwCode:   "myapp",
		Version:  "1.0.0",
		Filename: "myapp.zip",
		FileSize: 600 * 1024 * 1024,
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("Upload error = %v, want ErrFileTooLarge", err)
	}
}

func TestSoftwareUploadService_Upload_Duplicate(t *testing.T) {
	svc, db, _ := setupSoftwareUploadTest(t)

	data := []byte("sw data")
	in := SoftwareUploadInput{
		SwCode:   "dupapp",
		Version:  "1.0.0",
		Filename: "dup.zip",
		FileSize: int64(len(data)),
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	_, err = svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if !errors.Is(err, ErrVersionAlreadyExists) {
		t.Errorf("second Upload error = %v, want ErrVersionAlreadyExists", err)
	}

	var count int64
	db.Model(&model.SoftwareVersion{}).Count(&count)
	if count != 1 {
		t.Errorf("software version count = %d, want 1", count)
	}
}

func TestSoftwareUploadService_Upload_DuplicateWithPublished(t *testing.T) {
	svc, db, _ := setupSoftwareUploadTest(t)

	data := []byte("sw data")
	in := SoftwareUploadInput{
		SwCode:   "dupapp",
		Version:  "1.0.0",
		Filename: "dup.zip",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	db.Model(&model.SoftwareVersion{}).Where("id = ?", result.VersionId).Update("status", model.SoftwareStatusPublished)

	_, err = svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if !errors.Is(err, ErrVersionAlreadyExists) {
		t.Errorf("Upload with published duplicate error = %v, want ErrVersionAlreadyExists", err)
	}
}

func TestSoftwareUploadService_Upload_StatusIsDraft(t *testing.T) {
	svc, db, _ := setupSoftwareUploadTest(t)

	data := []byte("sw data")
	in := SoftwareUploadInput{
		SwCode:      "statusapp",
		Version:     "1.0.0",
		ReleaseNote: "check status",
		Filename:    "status.zip",
		FileSize:    int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if result.Status != "draft" {
		t.Errorf("Status = %q, want %q", result.Status, "draft")
	}

	var sw model.SoftwareVersion
	if err := db.Where("id = ?", result.VersionId).First(&sw).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if sw.Status != model.SoftwareStatusDraft {
		t.Errorf("DB Status = %v, want %v", sw.Status, model.SoftwareStatusDraft)
	}
	if sw.PublishedAt != nil {
		t.Error("PublishedAt must be nil for draft")
	}
}

func TestSoftwareUploadService_Upload_DifferentVersionNotDuplicate(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	data := []byte("sw data")

	in1 := SoftwareUploadInput{
		SwCode:   "myapp",
		Version:  "1.0.0",
		Filename: "myapp.zip",
		FileSize: int64(len(data)),
	}
	_, err := svc.Upload(context.Background(), in1, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	in2 := SoftwareUploadInput{
		SwCode:   "myapp",
		Version:  "2.0.0",
		Filename: "myapp.zip",
		FileSize: int64(len(data)),
	}
	result, err := svc.Upload(context.Background(), in2, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if err != nil {
		t.Fatalf("second Upload (different version) failed: %v", err)
	}
	if result.Status != string(model.SoftwareStatusDraft) {
		t.Errorf("Status = %q, want %q", result.Status, model.SoftwareStatusDraft)
	}
}

func TestSoftwareUploadService_Upload_ChecksumCorrect(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	data := []byte("software checksum test")
	in := SoftwareUploadInput{
		SwCode:   "ckapp",
		Version:  "1.0.0",
		Filename: "ck.zip",
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

func TestSoftwareUploadService_Upload_SemverFormats(t *testing.T) {
	svc, _, _ := setupSoftwareUploadTest(t)

	validVersions := []string{"1.0.0", "0.1.0", "10.20.30", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0+build.1", "1.0.0-alpha+build.1"}
	invalidVersions := []string{"1.0", "1", "v1.0.0", "1.0.0.0", "", "latest"}

	for _, v := range validVersions {
		t.Run("valid_"+v, func(t *testing.T) {
			data := []byte("test")
			in := SoftwareUploadInput{
				SwCode:   "semverapp",
				Version:  v,
				Filename: "test.zip",
				FileSize: int64(len(data)),
			}
			_, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
			if err != nil {
				t.Errorf("Upload with version=%q should succeed, got: %v", v, err)
			}
		})
	}

	for _, v := range invalidVersions {
		t.Run("invalid_"+v, func(t *testing.T) {
			in := SoftwareUploadInput{
				SwCode:   "semverapp2",
				Version:  v,
				Filename: "test.zip",
				FileSize: 100,
			}
			_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
			if !errors.Is(err, ErrInvalidParam) {
				t.Errorf("Upload with version=%q should fail with ErrInvalidParam, got: %v", v, err)
			}
		})
	}
}

var _ = strings.Repeat
