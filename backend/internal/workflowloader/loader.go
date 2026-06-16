package workflowloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

type LoadedDef struct {
	Name    string
	Version int
	File    string
}

type DefWarning struct {
	Name    string
	Version int
	File    string
	Message string
}

type DefError struct {
	File    string
	Message string
}

type LoadResult struct {
	Loaded   []LoadedDef
	Warnings []DefWarning
	Errors   []DefError
}

type Loader struct {
	dir     string
	defRepo repository.WorkflowDefRepo
	log     *logrus.Logger
}

func New(dir string, defRepo repository.WorkflowDefRepo, log *logrus.Logger) *Loader {
	return &Loader{dir: dir, defRepo: defRepo, log: log}
}

// Load scans dir for *.json files, validates, upserts. Returns result.
// Hard errors are in result.Errors; does NOT call os.Exit.
// Caller (main.go) calls os.Exit(1) on hard errors at startup.
func (l *Loader) Load(ctx context.Context) (LoadResult, error) {
	var result LoadResult

	entries, err := os.ReadDir(l.dir)
	if err != nil {
		if os.IsNotExist(err) {
			l.log.WithField("dir", l.dir).Info("workflows directory not found, skipping")
			return result, nil
		}
		return result, fmt.Errorf("read workflows dir %q: %w", l.dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(l.dir, entry.Name())
		loaded, warnings, defErr := l.loadFile(ctx, filePath)
		if defErr != nil {
			result.Errors = append(result.Errors, DefError{File: entry.Name(), Message: defErr.Error()})
			l.log.WithFields(logrus.Fields{"file": entry.Name(), "err": defErr}).Error("workflow def load failed")
			continue
		}
		for _, w := range warnings {
			result.Warnings = append(result.Warnings, DefWarning{
				Name: loaded.Name, Version: loaded.Version, File: entry.Name(), Message: w,
			})
			l.log.WithFields(logrus.Fields{"file": entry.Name(), "warning": w}).Warn("workflow def warning")
		}
		result.Loaded = append(result.Loaded, *loaded)
	}

	return result, nil
}

func (l *Loader) loadFile(ctx context.Context, filePath string) (*LoadedDef, []string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read file: %w", err)
	}

	var wfDef model.WorkflowDefinition
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&wfDef); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON: %w", err)
	}

	warnings, err := wfDef.Validate()
	if err != nil {
		return nil, nil, fmt.Errorf("validation: %w", err)
	}

	existing, err := l.defRepo.GetByNameVersion(ctx, wfDef.Name, wfDef.Version)
	if err != nil && err != repository.ErrNotFound {
		return nil, nil, fmt.Errorf("lookup existing: %w", err)
	}

	if existing != nil {
		existingJSON, _ := json.Marshal(existing.Definition)
		newJSON, _ := json.Marshal(wfDef)
		if string(existingJSON) == string(newJSON) {
			l.log.WithFields(logrus.Fields{"name": wfDef.Name, "version": wfDef.Version}).Info("workflow def unchanged, skipping")
			return &LoadedDef{Name: wfDef.Name, Version: wfDef.Version}, warnings, nil
		}
		return nil, nil, fmt.Errorf("version conflict: (%s, v%d) already exists with different content — bump version",
			wfDef.Name, wfDef.Version)
	}

	def := &model.WorkflowDef{
		ID:          uuid.New(),
		Name:        wfDef.Name,
		Version:     wfDef.Version,
		Description: wfDef.Description,
		Definition:  wfDef,
		CreatedAt:   time.Now().UTC(),
	}
	if err := l.defRepo.Create(ctx, def); err != nil {
		return nil, nil, fmt.Errorf("persist: %w", err)
	}

	l.log.WithFields(logrus.Fields{"name": wfDef.Name, "version": wfDef.Version}).Info("workflow def loaded")
	return &LoadedDef{Name: wfDef.Name, Version: wfDef.Version}, warnings, nil
}
