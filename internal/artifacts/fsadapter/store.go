package fsadapter

import (
	"os"
	"path/filepath"

	"shipwright/internal/artifacts/domain"
)

type Store struct {
	Layout domain.Layout
}

func NewStore(layout domain.Layout) Store {
	if layout.Root == "" {
		layout = domain.DefaultLayout()
	}
	return Store{Layout: layout}
}

func (s Store) Resolve(path string) string {
	return domain.ResolvePath(s.Layout, path)
}

func (s Store) Exists(path string) bool {
	info, err := os.Stat(s.Resolve(path))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (s Store) Write(path string, content string) error {
	resolved := s.Resolve(path)
	if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
		return err
	}
	return os.WriteFile(resolved, []byte(content), 0644)
}

func (s Store) Append(path string, content string) error {
	resolved := s.Resolve(path)
	if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func (s Store) CreateBaseStructure() error {
	for _, dir := range domain.AllDirs(s.Layout) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
