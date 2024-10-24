package settings

import (
	"downite/db"
	"downite/types"
	"encoding/json"
	"io"
	"os"
	"path"
	"runtime"
)

var (
	settingsFolderPath = ""
	dataFolderPath     = ""
)

var osSettingsFolderPaths = map[string]string{
	"windows": os.Getenv("AppData") + "\\downite",
	"linux":   os.Getenv("HOME") + "/.config/downite",
	"darwin":  os.Getenv("HOME") + "/.config/downite",
}

var osDataFolderPaths = map[string]string{
	"windows": os.Getenv("LocalAppData") + "\\downite",
	"linux":   os.Getenv("HOME") + "/.local/share/downite",
	"darwin":  os.Getenv("HOME") + "/.local/share/downite",
}

type DowniteSettingsSystem struct {
	Settings *types.DowniteSettings
}

func InitilizeSettingsSystem(db *db.Database, initialSettings *types.DowniteSettings) (*DowniteSettingsSystem, error) {
	settingsFolderPath = osSettingsFolderPaths[runtime.GOOS]
	dataFolderPath = osSettingsFolderPaths[runtime.GOOS]

	system := &DowniteSettingsSystem{}
	settingsFile, err := os.OpenFile(path.Join(settingsFolderPath, "settings.json"), os.O_RDWR, 0644)
	// if settings file doesn't exist create it
	if err != nil {
		if err.Error() == "no such file or directory" {
			defaultSettings := GetDefaultSettings()
			writeSettings(&defaultSettings)
			system.Settings = &defaultSettings
			return system, nil
		}
		return nil, err
	}
	// if settings file exists then read it

	foundSettings, err := readSettings(settingsFile)
	if err != nil {
		return nil, err
	}
	system.Settings = foundSettings
	return system, nil
}

func GetDefaultSettings() types.DowniteSettings {
	return types.DowniteSettings{
		Language:  "en",
		SavePaths: []string{},
	}
}

func writeSettings(settings *types.DowniteSettings) error {
	settingsFile, err := os.OpenFile(path.Join(settingsFolderPath, "settings.json"), os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	settingsJson, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = settingsFile.Write(settingsJson)
	if err != nil {
		return err
	}

	return nil
}

func readSettings(settingsFile *os.File) (settings *types.DowniteSettings, err error) {
	settingsFileBytes, err := io.ReadAll(settingsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(settingsFileBytes, settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (system *DowniteSettingsSystem) AddSavePath(savePath string) error {
	system.Settings.SavePaths = append(system.Settings.SavePaths, savePath)
	writeSettings(system.Settings)
	return nil
}

func (system *DowniteSettingsSystem) SetLanguage(language string) error {
	system.Settings.Language = language
	writeSettings(system.Settings)
	return nil
}
