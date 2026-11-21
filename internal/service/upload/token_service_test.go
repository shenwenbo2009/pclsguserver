package upload

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.UploadToken{}, &model.FileRecord{}, &model.Plugin{}, &model.SoftwareVersion{}, &model.AuditLog{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestUploadTokenService_Create(t *testing.T) {
	db := setupTestDB(t)
	tokenRepo := repo.NewUploadTokenRepo(db)
	svc := NewUploadTokenService(tokenRepo)

	token, record, err := svc.Create("test-token", "uploader1", []string{"plugin", "software"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if token == "" {
		t.Error("token should not be empty")
	}
	if len(token) < 10 {
		t.Errorf("token too short: %s", token)
	}

	if record == nil {
		t.Fatal("record should not be nil")
	}
	if record.ID == "" {
		t.Error("record ID should not be empty")
	}
	if record.TokenHash == "" {
		t.Error("TokenHash should not be empty")
	}
	if record.TokenHash == token {
		t.Error("TokenHash should not equal plaintext token (must be hashed)")
	}
	if record.Salt == "" {
		t.Error("Salt should not be empty")
	}
	if record.Status != model.UploadTokenStatusActive {
		t.Errorf("Status = %v, want %v", record.Status, model.UploadTokenStatusActive)
	}
	if len(record.Permissions) != 2 {
		t.Errorf("Permissions length = %d, want 2", len(record.Permissions))
	}
	if record.Name != "test-token" {
		t.Errorf("Name = %q, want %q", record.Name, "test-token")
	}
	if record.UploaderIdentity != "uploader1" {
		t.Errorf("UploaderIdentity = %q, want %q", record.UploaderIdentity, "uploader1")
	}
}

func TestUploadTokenService_Validate(t *testing.T) {
	db := setupTestDB(t)
	tokenRepo := repo.NewUploadTokenRepo(db)
	svc := NewUploadTokenService(tokenRepo)

	token, _, err := svc.Create("test-token", "uploader1", []string{"plugin"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		got, err := svc.Validate(token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if got == nil {
			t.Fatal("Validate returned nil")
		}
		if got.Name != "test-token" {
			t.Errorf("Name = %q, want %q", got.Name, "test-token")
		}
		if got.Status != model.UploadTokenStatusActive {
			t.Errorf("Status = %v, want %v", got.Status, model.UploadTokenStatusActive)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := svc.Validate("")
		if err == nil {
			t.Error("Validate with empty token should return error")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := svc.Validate("utk_invalidtoken")
		if err == nil {
			t.Error("Validate with invalid token should return error")
		}
	})

	t.Run("wrong token format", func(t *testing.T) {
		_, err := svc.Validate("wrong_format_token")
		if err == nil {
			t.Error("Validate with wrong format should return error")
		}
	})
}

func TestUploadTokenService_Disable(t *testing.T) {
	db := setupTestDB(t)
	tokenRepo := repo.NewUploadTokenRepo(db)
	svc := NewUploadTokenService(tokenRepo)

	token, record, err := svc.Create("test-token", "uploader1", []string{"plugin"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("disable existing token", func(t *testing.T) {
		err := svc.Disable(record.ID)
		if err != nil {
			t.Fatalf("Disable failed: %v", err)
		}

		got, err := svc.Validate(token)
		if err != nil {
			t.Fatalf("Validate after disable failed: %v", err)
		}
		if got.Status != model.UploadTokenStatusDisabled {
			t.Errorf("Status = %v, want %v", got.Status, model.UploadTokenStatusDisabled)
		}
		if got.DisabledAt == nil {
			t.Error("DisabledAt should not be nil after disable")
		}
	})

	t.Run("disable non-existent token", func(t *testing.T) {
		err := svc.Disable("non-existent-id")
		if err == nil {
			t.Error("Disable non-existent token should return error")
		}
	})
}

func TestUploadTokenService_List(t *testing.T) {
	db := setupTestDB(t)
	tokenRepo := repo.NewUploadTokenRepo(db)
	svc := NewUploadTokenService(tokenRepo)

	for i := 0; i < 3; i++ {
		_, _, err := svc.Create("test-token-"+string(rune('A'+i)), "uploader1", []string{"plugin"})
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
	}

	tokens, total, err := svc.List(1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(tokens) != 3 {
		t.Errorf("tokens length = %d, want 3", len(tokens))
	}

	t.Run("pagination", func(t *testing.T) {
		tokens, total, err := svc.List(1, 2)
		if err != nil {
			t.Fatalf("List page 1 failed: %v", err)
		}
		if total != 3 {
			t.Errorf("total = %d, want 3", total)
		}
		if len(tokens) != 2 {
			t.Errorf("tokens length = %d, want 2", len(tokens))
		}

		tokens2, _, err := svc.List(2, 2)
		if err != nil {
			t.Fatalf("List page 2 failed: %v", err)
		}
		if len(tokens2) != 1 {
			t.Errorf("tokens2 length = %d, want 1", len(tokens2))
		}
	})
}

func TestUploadTokenService_TokenHashNotPlaintext(t *testing.T) {
	db := setupTestDB(t)
	tokenRepo := repo.NewUploadTokenRepo(db)
	svc := NewUploadTokenService(tokenRepo)

	token, record, err := svc.Create("test-token", "uploader1", []string{"plugin"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if record.TokenHash == token {
		t.Error("TokenHash must not equal plaintext token")
	}

	var stored model.UploadToken
	if err := db.Where("id = ?", record.ID).First(&stored).Error; err != nil {
		t.Fatalf("query stored token failed: %v", err)
	}

	if stored.TokenHash == token {
		t.Error("stored TokenHash must not equal plaintext token")
	}
	if stored.Salt == "" {
		t.Error("stored Salt must not be empty")
	}
}
