package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Texts represents all translatable strings
type Texts map[string]string

// I18n handles internationalization
type I18n struct {
	translations map[string]Texts
	logger       *logrus.Entry
}

// NewI18n creates a new internationalization manager
func NewI18n() *I18n {
	return &I18n{
		translations: make(map[string]Texts),
		logger:       logrus.WithField("component", "i18n"),
	}
}

// LoadTranslations loads translation files from the specified directory
func (i *I18n) LoadTranslations(directory string) error {
	i.logger.WithField("directory", directory).Info("Loading translation files")

	languages := []string{"en", "ru"}

	for _, lang := range languages {
		if err := i.loadLanguage(directory, lang); err != nil {
			return fmt.Errorf("failed to load language %s: %w", lang, err)
		}
	}

	i.logger.WithField("languages", len(i.translations)).Info("All translations loaded successfully")
	return nil
}

// loadLanguage loads a single language file
func (i *I18n) loadLanguage(directory, language string) error {
	filePath := filepath.Join(directory, language+".json")

	i.logger.WithFields(logrus.Fields{
		"language": language,
		"path":     filePath,
	}).Debug("Loading translation file")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read translation file: %w", err)
	}

	var texts Texts
	if err := json.Unmarshal(data, &texts); err != nil {
		return fmt.Errorf("failed to parse translation JSON: %w", err)
	}

	i.translations[language] = texts

	i.logger.WithField("language", language).Info("Translation loaded successfully")
	return nil
}

// GetTexts returns translations for the specified language
func (i *I18n) GetTexts(language string) Texts {
	if texts, exists := i.translations[language]; exists {
		return texts
	}

	// Fallback to English if requested language is not available
	if texts, exists := i.translations["en"]; exists {
		i.logger.WithField("requested_language", language).Warn("Language not found, falling back to English")
		return texts
	}

	// Return empty texts if nothing is available
	i.logger.WithField("requested_language", language).Error("No translations available")
	return make(Texts)
}

// GetSupportedLanguages returns list of supported languages
func (i *I18n) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		languages = append(languages, lang)
	}
	return languages
}

// DetectLanguage detects language from query parameter or returns default
func DetectLanguage(langParam string) string {
	supportedLanguages := map[string]bool{
		"en": true,
		"ru": true,
	}

	if langParam != "" && supportedLanguages[langParam] {
		return langParam
	}

	return "en" // Default language
}
