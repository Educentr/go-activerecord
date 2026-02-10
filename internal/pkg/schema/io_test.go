package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		wantNil   bool
		wantErr   bool
		errSubstr string
	}{
		{
			name: "file not found returns nil nil",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "valid JSON returns parsed Schema",
			setup: func() string {
				path := filepath.Join(tmpDir, "valid.json")
				data := `{"version":"1.0","backend":"postgres","tables":{"users":{"name":"users","columns":[],"indexes":[]}}}`
				if err := os.WriteFile(path, []byte(data), 0644); err != nil {
					t.Fatal(err)
				}

				return path
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "invalid JSON returns error",
			setup: func() string {
				path := filepath.Join(tmpDir, "invalid.json")
				if err := os.WriteFile(path, []byte("{bad json}"), 0644); err != nil {
					t.Fatal(err)
				}

				return path
			},
			wantNil:   true,
			wantErr:   true,
			errSubstr: "parse schema JSON",
		},
		{
			name: "read error dir as file",
			setup: func() string {
				dir := filepath.Join(tmpDir, "adir")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			wantNil:   true,
			wantErr:   true,
			errSubstr: "read schema file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			got, err := LoadSchema(path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil && got != nil {
				t.Errorf("expected nil schema, got %+v", got)
			}

			if !tt.wantNil && got == nil {
				t.Error("expected non-nil schema, got nil")
			}
		})
	}
}

func TestLoadSchema_ValidContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schema.json")
	data := `{"version":"1.0","backend":"postgres","tables":{"users":{"name":"users","backend":"postgres","columns":[{"name":"id","type":"BIGINT","go_type":"int64","position":0}],"indexes":[]}}}`

	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSchema(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Version != "1.0" {
		t.Errorf("expected version '1.0', got '%s'", s.Version)
	}

	if s.Backend != "postgres" {
		t.Errorf("expected backend 'postgres', got '%s'", s.Backend)
	}

	tbl, ok := s.Tables["users"]
	if !ok {
		t.Fatal("expected 'users' table")
	}

	if len(tbl.Columns) != 1 {
		t.Errorf("expected 1 column, got %d", len(tbl.Columns))
	}
}

func TestSaveSchema(t *testing.T) {
	t.Run("creates dir and writes JSON roundtrip", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "sub", "schema.json")

		s := NewSchema("postgres")
		s.Tables["users"] = Table{
			Name:    "users",
			Backend: "postgres",
			Columns: []Column{{Name: "id", Type: "BIGINT", GoType: "int64"}},
		}

		if err := SaveSchema(s, path); err != nil {
			t.Fatalf("SaveSchema error: %v", err)
		}

		loaded, err := LoadSchema(path)
		if err != nil {
			t.Fatalf("LoadSchema error: %v", err)
		}

		if loaded.Backend != "postgres" {
			t.Errorf("expected backend 'postgres', got '%s'", loaded.Backend)
		}

		if _, ok := loaded.Tables["users"]; !ok {
			t.Error("expected 'users' table after roundtrip")
		}
	})

	t.Run("overwrites existing", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "schema.json")

		s1 := NewSchema("postgres")
		s1.Tables["v1"] = Table{Name: "v1"}

		if err := SaveSchema(s1, path); err != nil {
			t.Fatalf("first save error: %v", err)
		}

		s2 := NewSchema("postgres")
		s2.Tables["v2"] = Table{Name: "v2"}

		if err := SaveSchema(s2, path); err != nil {
			t.Fatalf("second save error: %v", err)
		}

		loaded, err := LoadSchema(path)
		if err != nil {
			t.Fatalf("LoadSchema error: %v", err)
		}

		if _, ok := loaded.Tables["v1"]; ok {
			t.Error("v1 table should not exist after overwrite")
		}

		if _, ok := loaded.Tables["v2"]; !ok {
			t.Error("v2 table should exist after overwrite")
		}
	})

	t.Run("dir creation fails", func(t *testing.T) {
		path := "/dev/null/sub/schema.json"

		err := SaveSchema(NewSchema("postgres"), path)
		if err == nil {
			t.Fatal("expected error when dir creation fails")
		}

		if !strings.Contains(err.Error(), "create directory") {
			t.Errorf("error %q should contain 'create directory'", err.Error())
		}
	})
}

func TestSaveDDL(t *testing.T) {
	t.Run("writes to new dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "newdir", "schema.sql")
		content := []byte("CREATE TABLE users (id BIGINT);")

		if err := SaveDDL(content, path); err != nil {
			t.Fatalf("SaveDDL error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("content mismatch: got %q, want %q", string(data), string(content))
		}
	})

	t.Run("dir creation error", func(t *testing.T) {
		err := SaveDDL([]byte("test"), "/dev/null/sub/schema.sql")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSaveMigration(t *testing.T) {
	t.Run("creates migrations dir and file", func(t *testing.T) {
		tmpDir := t.TempDir()

		migration := &Migration{
			Number:      1,
			Description: "test migration",
			Up:          []string{"CREATE TABLE t1 (id INT)"},
			Down:        []string{"DROP TABLE t1"},
		}

		if err := SaveMigration(migration, tmpDir); err != nil {
			t.Fatalf("SaveMigration error: %v", err)
		}

		expectedPath := filepath.Join(tmpDir, "migrations", migration.Filename())

		data, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}

		expected := migration.Format()
		if string(data) != expected {
			t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(data), expected)
		}
	})

	t.Run("dir creation error", func(t *testing.T) {
		migration := &Migration{Number: 1, Description: "test", Up: []string{"UP"}, Down: []string{"DOWN"}}

		err := SaveMigration(migration, "/dev/null/bad")
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "create migrations directory") {
			t.Errorf("error %q should contain 'create migrations directory'", err.Error())
		}
	})
}

func TestGetNextMigrationNumber(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    int
		wantErr bool
	}{
		{
			name: "no migrations dir",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: 1,
		},
		{
			name: "empty dir",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				if err := os.MkdirAll(filepath.Join(dir, "migrations"), 0755); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: 1,
		},
		{
			name: "single file 001_init.sql",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				migrDir := filepath.Join(dir, "migrations")

				if err := os.MkdirAll(migrDir, 0755); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(migrDir, "001_init.sql"), []byte(""), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: 2,
		},
		{
			name: "non-sequential 001 003 002",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				migrDir := filepath.Join(dir, "migrations")

				if err := os.MkdirAll(migrDir, 0755); err != nil {
					t.Fatal(err)
				}

				for _, name := range []string{"001_init.sql", "003_third.sql", "002_second.sql"} {
					if err := os.WriteFile(filepath.Join(migrDir, name), []byte(""), 0644); err != nil {
						t.Fatal(err)
					}
				}

				return dir
			},
			want: 4,
		},
		{
			name: "ignores non-matching files",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				migrDir := filepath.Join(dir, "migrations")

				if err := os.MkdirAll(migrDir, 0755); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(migrDir, "README.md"), []byte(""), 0644); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(migrDir, "005_real.sql"), []byte(""), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: 6,
		},
		{
			name: "ignores subdirectories",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				migrDir := filepath.Join(dir, "migrations")

				if err := os.MkdirAll(filepath.Join(migrDir, "010_subdir"), 0755); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			want: 1,
		},
		{
			name: "read error migrations is a file",
			setup: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "migrations"), []byte("not a dir"), 0644); err != nil {
					t.Fatal(err)
				}

				return dir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			got, err := GetNextMigrationNumber(dir)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("GetNextMigrationNumber() = %d, want %d", got, tt.want)
			}
		})
	}
}
