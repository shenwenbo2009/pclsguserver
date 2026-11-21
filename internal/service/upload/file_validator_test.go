package upload

import (
	"testing"
)

func TestValidatePackageType(t *testing.T) {
	v := NewFileValidator(
		[]string{".zip", ".tar.gz", ".tgz", ".7z", ".rar"},
		[]string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"},
	)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"zip file", "plugin.zip", false},
		{"tar.gz file", "plugin.tar.gz", false},
		{"tgz file", "plugin.tgz", false},
		{"7z file", "plugin.7z", false},
		{"rar file", "plugin.rar", false},
		{"uppercase ZIP", "PLUGIN.ZIP", false},
		{"uppercase TAR.GZ", "PLUGIN.TAR.GZ", false},
		{"exe file", "plugin.exe", true},
		{"sh file", "plugin.sh", true},
		{"no extension", "plugin", true},
		{"js file", "plugin.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePackageType(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePackageType(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileType(t *testing.T) {
	v := NewFileValidator(
		[]string{".zip", ".tar.gz", ".tgz", ".7z", ".rar"},
		[]string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"},
	)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"jpg file", "photo.jpg", false},
		{"jpeg file", "photo.jpeg", false},
		{"png file", "photo.png", false},
		{"gif file", "photo.gif", false},
		{"webp file", "photo.webp", false},
		{"pdf file", "doc.pdf", false},
		{"doc file", "doc.doc", false},
		{"docx file", "doc.docx", false},
		{"txt file", "doc.txt", false},
		{"csv file", "data.csv", false},
		{"xlsx file", "data.xlsx", false},
		{"uppercase JPG", "PHOTO.JPG", false},
		{"uppercase PNG", "PHOTO.PNG", false},
		{"exe file", "photo.exe", true},
		{"sh file", "photo.sh", true},
		{"zip file (not allowed for general)", "photo.zip", true},
		{"no extension", "photo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFileType(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileType(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSize(t *testing.T) {
	v := NewFileValidator(nil, nil)

	tests := []struct {
		name    string
		size    int64
		maxSize int64
		wantErr bool
	}{
		{"within limit", 100, 200, false},
		{"exactly at limit", 200, 200, false},
		{"exceeds limit", 201, 200, true},
		{"zero size", 0, 200, false},
		{"large within limit", 104857600, 104857600, false},
		{"large exceeds limit", 104857601, 104857600, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSize(tt.size, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSize(%d, %d) error = %v, wantErr %v", tt.size, tt.maxSize, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	v := NewFileValidator(nil, nil)

	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{"simple filename", "photo.jpg", "photo.jpg", false},
		{"path traversal", "../../etc/passwd", "passwd", false},
		{"windows path traversal", "..\\..\\windows\\system32", "system32", false},
		{"with slash", "dir/file.txt", "file.txt", false},
		{"with backslash", "dir\\file.txt", "file.txt", false},
		{"with percent", "file%20name.txt", "file20name.txt", false},
		{"double dot only", "..", "", true},
		{"dot only", ".", "", true},
		{"empty", "", "", true},
		{"normal with dots in name", "my.file.name.txt", "my.file.name.txt", false},
		{"leading dots", "...hidden.txt", ".hidden.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.SanitizeFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilenameDirectoryTraversal(t *testing.T) {
	v := NewFileValidator(nil, nil)

	dangerousNames := []string{
		"../../etc/passwd",
		"../../../root/.ssh/id_rsa",
		"..\\..\\windows\\system32\\config",
		"./../../secret",
	}

	for _, name := range dangerousNames {
		t.Run(name, func(t *testing.T) {
			clean, err := v.SanitizeFilename(name)
			if err != nil {
				return
			}
			if clean == "" || clean == "." {
				t.Errorf("SanitizeFilename(%q) returned empty/dot", name)
			}
			if contains(clean, "..") {
				t.Errorf("SanitizeFilename(%q) still contains '..': %q", name, clean)
			}
			if contains(clean, "/") {
				t.Errorf("SanitizeFilename(%q) still contains '/': %q", name, clean)
			}
			if contains(clean, "\\") {
				t.Errorf("SanitizeFilename(%q) still contains '\\': %q", name, clean)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
