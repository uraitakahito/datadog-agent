// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package integration creates a launcher to track logs from integrations
package integration

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/filesystem"
	ddLog "github.com/DataDog/datadog-agent/pkg/util/log"

	"github.com/DataDog/datadog-agent/comp/logs/agent/config"
	integrations "github.com/DataDog/datadog-agent/comp/logs/integrations/def"
	pkgConfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/launchers"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/schedulers/ad"
	"github.com/DataDog/datadog-agent/pkg/logs/sources"
	"github.com/DataDog/datadog-agent/pkg/logs/tailers"
)

var endOfLine = []byte{'\n'}

// Launcher checks for launcher integrations, creates files for integrations to
// write logs to, then creates file sources for the file launcher to tail
type Launcher struct {
	sources                   *sources.LogSources
	addedConfigs              chan integrations.IntegrationConfig
	stop                      chan struct{}
	runPath                   string
	integrationsLogsChan      chan integrations.IntegrationLog
	integrationToFile         map[string]*FileInfo
	fileSizeMax               int64
	combinedUsageMax          int64
	combinedUsageSize         int64
	leastRecentlyModifiedTime time.Time
	leastRecentlyModifiedFile *FileInfo
	// writeLogToFile is used as a function pointer, so it can be overridden in
	// testing to make deterministic tests
	writeLogToFileFunction func(filepath, log string) error
}

// FileInfo stores information about each file that is needed in order to keep
// track of the combined and overall disk usage by the logs files
type FileInfo struct {
	filename     string
	lastModified time.Time
	size         int64
}

// NewLauncher creates and returns an integrations launcher, and creates the
// path for integrations files to run in
func NewLauncher(sources *sources.LogSources, integrationsLogsComp integrations.Component) *Launcher {
	runPath := filepath.Join(pkgConfig.Datadog().GetString("logs_config.run_path"), "integrations")
	err := os.MkdirAll(runPath, 0755)
	if err != nil {
		ddLog.Warn("Unable to make integrations logs directory: ", err)
		return nil
	}

	logsTotalUsageSetting := pkgConfig.Datadog().GetInt64("logs_config.integrations_logs_total_usage") * 1024 * 1024
	logsUsageRatio := pkgConfig.Datadog().GetFloat64("logs_config.integrations_logs_disk_ratio")
	maxDiskUsage, err := computeMaxDiskUsage(runPath, logsTotalUsageSetting, logsUsageRatio)
	if err != nil {
		ddLog.Warn("Unable to compute integrations logs max disk usage, defaulting to set value: ", err)
		maxDiskUsage = logsTotalUsageSetting
	}

	return &Launcher{
		sources:              sources,
		runPath:              runPath,
		fileSizeMax:          pkgConfig.Datadog().GetInt64("logs_config.integrations_logs_files_max_size") * 1024 * 1024,
		combinedUsageMax:     maxDiskUsage,
		combinedUsageSize:    0,
		stop:                 make(chan struct{}),
		integrationsLogsChan: integrationsLogsComp.Subscribe(),
		addedConfigs:         integrationsLogsComp.SubscribeIntegration(),
		integrationToFile:    make(map[string]*FileInfo),
		// Set the initial least recently modified time to the largest possible
		// value, used for the first comparison
		leastRecentlyModifiedTime: time.Unix(1<<63-62135596801, 999999999),
		writeLogToFileFunction:    writeLogToFile,
	}
}

// Start starts the launcher and launches the run loop in a go function
func (s *Launcher) Start(_ launchers.SourceProvider, _ pipeline.Provider, _ auditor.Registry, _ *tailers.TailerTracker) {
	err := s.scanInitialFiles(s.runPath)
	if err != nil {
		ddLog.Warn("Unable to scan existing log files: ", err)
	}

	go s.run()
}

// Stop stops the launcher
func (s *Launcher) Stop() {
	s.stop <- struct{}{}
}

// run checks if there are new files to tail and tails them
func (s *Launcher) run() {
	for {
		select {
		case cfg := <-s.addedConfigs:
			sources, err := ad.CreateSources(cfg.Config)
			if err != nil {
				ddLog.Warn("Failed to create source ", err)
				continue
			}

			for _, source := range sources {
				// TODO: integrations should only be allowed to have one IntegrationType config.
				if source.Config.Type == config.IntegrationType {
					logFile, err := s.createFile(cfg.IntegrationID)
					if err != nil {
						ddLog.Warn("Failed to create integration log file: ", err)
						continue
					}
					filetypeSource := s.makeFileSource(source, logFile.filename)
					s.sources.AddSource(filetypeSource)

					// file to write the incoming logs to
					s.integrationToFile[cfg.IntegrationID] = logFile
				}
			}

		case log := <-s.integrationsLogsChan:
			s.receiveLogs(log)
		case <-s.stop:
			return
		}
	}
}

// receiveLogs handles writing incoming logs to their respective file as well as
// enforcing size limitations
func (s *Launcher) receiveLogs(log integrations.IntegrationLog) {
	fileToUpdate := s.integrationToFile[log.IntegrationID]

	// Ensure the individual file doesn't exceed integrations_logs_files_max_size
	// Add 1 because we write the \n at the end as well
	logSize := int64(len(log.Log)) + 1
	if fileToUpdate.size+logSize > s.fileSizeMax {
		s.combinedUsageSize -= fileToUpdate.size
		err := s.deleteAndRemakeFile(fileToUpdate.filename)
		if err != nil {
			ddLog.Warn("Failed to delete and remake oversize file", err)
			return
		}

		fileToUpdate.size = 0
	}

	// Ensure combined logs usage doesn't exceed integrations_logs_total_usage
	for s.combinedUsageSize+logSize > s.combinedUsageMax {
		s.deleteLeastRecentlyModified()
		s.updateLeastRecentlyModifiedFile()
	}

	err := s.writeLogToFileFunction(filepath.Join(s.runPath, fileToUpdate.filename), log.Log)
	if err != nil {
		ddLog.Warn("Error writing log to file: ", err)
	}
	s.combinedUsageSize += logSize

	// Update information for the modified file
	fileToUpdate.lastModified = time.Now()
	fileToUpdate.size += logSize

	// Update leastRecentlyModifiedFile
	if s.leastRecentlyModifiedFile == fileToUpdate && len(s.integrationToFile) > 1 {
		s.updateLeastRecentlyModifiedFile()
	}

}

// deleteLeastRecentlyModified deletes and remakes the least recently log file
func (s *Launcher) deleteLeastRecentlyModified() {
	s.combinedUsageSize -= s.leastRecentlyModifiedFile.size
	err := s.deleteAndRemakeFile(filepath.Join(s.runPath, s.leastRecentlyModifiedFile.filename))
	if err != nil {
		ddLog.Warn("Unable to delete and remake least recently modified file: ", err)
	}

	s.leastRecentlyModifiedFile.size = 0
	s.leastRecentlyModifiedFile.lastModified = time.Now()
}

// updateLeastRecentlyModifiedFile finds the least recently modified file among
// all the files tracked by the integrations launcher and sets the
// leastRecentlyModifiedFile to that file
func (s *Launcher) updateLeastRecentlyModifiedFile() {
	leastRecentlyModifiedTime := time.Now()
	var leastRecentlyModifiedFile *FileInfo

	for _, fileInfo := range s.integrationToFile {
		if fileInfo.lastModified.Before(leastRecentlyModifiedTime) {
			leastRecentlyModifiedFile = fileInfo
			leastRecentlyModifiedTime = fileInfo.lastModified
		}
	}

	s.leastRecentlyModifiedFile = leastRecentlyModifiedFile
	s.leastRecentlyModifiedTime = leastRecentlyModifiedTime
}

// writeLogToFile is used as a function pointer that writes a log to a given file
func writeLogToFile(logFilePath, log string) error {
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		ddLog.Warn("Failed to open file to write log to: ", err)
		return err
	}

	defer file.Close()

	_, err = file.WriteString(log)
	if err != nil {
		ddLog.Warn("Failed to write integration log to file: ", err)
		return err
	}
	if _, err = file.Write(endOfLine); err != nil {
		ddLog.Warn("Failed to write integration log to file: ", err)
		return err
	}
	return nil
}

// makeFileSource Turns an integrations source into a logsSource
func (s *Launcher) makeFileSource(source *sources.LogSource, logFilePath string) *sources.LogSource {
	fileSource := sources.NewLogSource(source.Name, &config.LogsConfig{
		Type:        config.FileType,
		TailingMode: source.Config.TailingMode,
		Path:        logFilePath,
		Name:        source.Config.Name,
		Source:      source.Config.Source,
		Service:     source.Config.Service,
		Tags:        source.Config.Tags,
	})

	fileSource.SetSourceType(sources.IntegrationSourceType)
	return fileSource
}

// TODO Change file naming to reflect ID once logs from go interfaces gets merged.
// createFile creates a file for the logsource
func (s *Launcher) createFile(source string) (*FileInfo, error) {
	filepath := s.integrationLogFilePath(source)

	file, err := os.Create(filepath)
	if err != nil {
		return nil, nil
	}
	ddLog.Info("Successfully created integrations log file.")
	defer file.Close()

	fileInfo := &FileInfo{
		filename:     filepath,
		lastModified: time.Now(),
		size:         0,
	}

	return fileInfo, nil
}

// integrationLoglogFilePath returns a file path to use for an integration log file
func (s *Launcher) integrationLogFilePath(id string) string {
	fileName := strings.ReplaceAll(id, ":", "_") + ".log"
	logFilePath := filepath.Join(s.runPath, fileName)

	return logFilePath
}

// deleteAndRemakeFile deletes log files and creates a new empty file with the
// same name
func (s *Launcher) deleteAndRemakeFile(filepath string) error {
	err := os.Remove(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			ddLog.Warn("File does not exist, creating new one: ", err)
		} else {
			ddLog.Warn("Error deleting file: ", err)
		}
	} else {
		ddLog.Info("Successfully deleted oversize log file, creating new one.")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

// computerDiskUsageMax computes the max disk space the launcher can use based
// off the integrations_logs_disk_ratio and integrations_logs_total_usage
// settings
func computeMaxDiskUsage(runPath string, logsTotalUsageSetting int64, usageRatio float64) (int64, error) {
	usage, err := filesystem.NewDisk().GetUsage(runPath)
	if err != nil {
		return 0, err
	}

	diskReserved := float64(usage.Total) * (1 - usageRatio)
	diskAvailable := int64(usage.Available) - int64(math.Ceil(diskReserved))

	return min(logsTotalUsageSetting, diskAvailable), nil
}

// scanInitialFiles scans the run path for initial files and then adds them to
// be managed by the launcher
func (s *Launcher) scanInitialFiles(dir string) error {
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			fileInfo := &FileInfo{
				filename:     info.Name(),
				size:         info.Size(),
				lastModified: info.ModTime(),
			}

			integrationID := fileNameToID(fileInfo.filename)

			s.integrationToFile[integrationID] = fileInfo

			if s.leastRecentlyModifiedFile == nil {
				s.leastRecentlyModifiedFile = fileInfo
			} else if fileInfo.lastModified.Before(s.leastRecentlyModifiedFile.lastModified) {
				s.leastRecentlyModifiedFile = fileInfo
			}

			for s.combinedUsageSize+info.Size() > s.combinedUsageMax {
				s.deleteLeastRecentlyModified()
				s.updateLeastRecentlyModifiedFile()
			}

			s.combinedUsageSize += info.Size()
		}

		return nil
	})

	return err
}

func fileNameToID(fileName string) string {
	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	lastUnderscoreIndex := strings.LastIndex(baseName, "_")
	if lastUnderscoreIndex == -1 {
		return baseName
	}

	integrationID := baseName[:lastUnderscoreIndex] + ":" + baseName[lastUnderscoreIndex+1:]
	return integrationID
}
