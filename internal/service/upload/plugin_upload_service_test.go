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

func setupPluginUploadTest(t *testing.T) (*PluginUploadService, *gorm.DB, *store.FsStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.Plugin{}, &model.AuditLog{}); err != nil {
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

	pluginRepo := repo.NewPluginRepo(db)
	auditRepo := repo.NewAuditRepo(db)
	auditSvc := audit.NewService(auditRepo)

	svc := NewPluginUploadService(validator, fsStore, pluginRepo, auditSvc, db, 200*1024*1024)
	return svc, db, fsStore
}

func TestPluginUploadService_Upload_Success(t *testing.T) {
	svc, db, _ := setupPluginUploadTest(t)

	data := bytes.Repeat([]byte("a"), 1024)
	in := PluginUploadInput{
		Name:        "my-plugin",
		Version:     "1.0.0",
		Author:      "author1",
		Category:    "utility",
		Description: "a test plugin",
		Keywords:    []string{"test", "demo"},
		Filename:    "my-plugin.zip",
		FileSize:    int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.PluginId == "" {
		t.Error("PluginId should not be empty")
	}
	if result.Status != string(model.PluginStatusReviewing) {
		t.Errorf("Status = %q, want %q", result.Status, model.PluginStatusReviewing)
	}
	if result.Checksum == "" {
		t.Error("Checksum should not be empty")
	}
	if len(result.Checksum) != 64 {
		t.Errorf("Checksum length = %d, want 64 (SHA256 hex)", len(result.Checksum))
	}
	if result.Size != int64(len(data)) {
		t.Errorf("Size = %d, want %d", result.Size, len(data))
	}
	if result.UploadedAt == "" {
		t.Error("UploadedAt should not be empty")
	}

	var plugin model.Plugin
	if err := db.Where("id = ?", result.PluginId).First(&plugin).Error; err != nil {
		t.Fatalf("query plugin failed: %v", err)
	}
	if plugin.Status != model.PluginStatusReviewing {
		t.Errorf("plugin Status = %v, want %v", plugin.Status, model.PluginStatusReviewing)
	}
	if plugin.SubmitterId != "uploader1" {
		t.Errorf("SubmitterId = %q, want %q", plugin.SubmitterId, "uploader1")
	}
	if plugin.Name != "my-plugin" {
		t.Errorf("Name = %q, want %q", plugin.Name, "my-plugin")
	}

	var logs []model.AuditLog
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("query audit logs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs count = %d, want 1", len(logs))
	}
	if logs[0].Action != model.AuditActionPluginUploaded {
		t.Errorf("audit Action = %q, want %q", logs[0].Action, model.AuditActionPluginUploaded)
	}
}

func TestPluginUploadService_Upload_MetadataValidation(t *testing.T) {
	svc, _, _ := setupPluginUploadTest(t)

	tests := []struct {
		name string
		in   PluginUploadInput
		err  error
	}{
		{
			name: "name too long",
			in: PluginUploadInput{
				Name:     strings.Repeat("x", 101),
				Version:  "1.0.0",
				Author:   "author1",
				Category: "utility",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "name empty",
			in: PluginUploadInput{
				Name:     "",
				Version:  "1.0.0",
				Author:   "author1",
				Category: "utility",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "invalid version (not semver)",
			in: PluginUploadInput{
				Name:     "plugin1",
				Version:  "1.0",
				Author:   "author1",
				Category: "utility",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "author too long",
			in: PluginUploadInput{
				Name:     "plugin1",
				Version:  "1.0.0",
				Author:   strings.Repeat("x", 101),
				Category: "utility",
				Filename: "test.zip",
				FileSize: 100,
			},
			err: ErrInvalidParam,
		},
		{
			name: "category empty",
			in: PluginUploadInput{
				Name:     "plugin1",
				Version:  "1.0.0",
				Author:   "author1",
				Category: "",
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

func TestPluginUploadService_Upload_InvalidFileType(t *testing.T) {
	svc, _, _ := setupPluginUploadTest(t)

	in := PluginUploadInput{
		Name:     "plugin1",
		Version:  "1.0.0",
		Author:   "author1",
		Category: "utility",
		Filename: "plugin.exe",
		FileSize: 100,
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
	if !errors.Is(err, ErrInvalidFileType) {
		t.Errorf("Upload error = %v, want ErrInvalidFileType", err)
	}
}

func TestPluginUploadService_Upload_FileTooLarge(t *testing.T) {
	svc, _, _ := setupPluginUploadTest(t)

	in := PluginUploadInput{
		Name:     "plugin1",
		Version:  "1.0.0",
		Author:   "author1",
		Category: "utility",
		Filename: "plugin.zip",
		FileSize: 300 * 1024 * 1024,
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader([]byte("data")), "uploader1", "req-1", "127.0.0.1")
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("Upload error = %v, want ErrFileTooLarge", err)
	}
}

func TestPluginUploadService_Upload_Duplicate(t *testing.T) {
	svc, db, _ := setupPluginUploadTest(t)

	data := []byte("plugin data")
	in := PluginUploadInput{
		Name:     "dup-plugin",
		Version:  "1.0.0",
		Author:   "author1",
		Category: "utility",
		Filename: "dup.zip",
		FileSize: int64(len(data)),
	}

	_, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	_, err = svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if !errors.Is(err, ErrPluginAlreadyExists) {
		t.Errorf("second Upload error = %v, want ErrPluginAlreadyExists", err)
	}

	var count int64
	db.Model(&model.Plugin{}).Count(&count)
	if count != 1 {
		t.Errorf("plugin count = %d, want 1", count)
	}
}

func TestPluginUploadService_Upload_DuplicateWithPublished(t *testing.T) {
	svc, db, _ := setupPluginUploadTest(t)

	data := []byte("plugin data")
	in := PluginUploadInput{
		Name:     "dup-plugin",
		Version:  "1.0.0",
		Author:   "author1",
		Category: "utility",
		Filename: "dup.zip",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("first Upload failed: %v", err)
	}

	db.Model(&model.Plugin{}).Where("id = ?", result.PluginId).Update("status", model.PluginStatusPublished)

	_, err = svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-2", "127.0.0.1")
	if !errors.Is(err, ErrPluginAlreadyExists) {
		t.Errorf("Upload with published duplicate error = %v, want ErrPluginAlreadyExists", err)
	}
}

func TestPluginUploadService_Upload_ChecksumCorrect(t *testing.T) {
	svc, _, _ := setupPluginUploadTest(t)

	data := []byte("checksum test data")
	in := PluginUploadInput{
		Name:     "checksum-plugin",
		Version:  "1.0.0",
		Author:   "author1",
		Category: "utility",
		Filename: "plugin.zip",
		FileSize: int64(len(data)),
	}

	result, err := svc.Upload(context.Background(), in, bytes.NewReader(data), "uploader1", "req-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	expectedChecksum := computeExpectedChecksum(t, data)
	if result.Checksum != expectedChecksum {
		t.Errorf("Checksum = %q, want %q", result.Checksum, expectedChecksum)
	}
}
