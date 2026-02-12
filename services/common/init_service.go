package common

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"gorm.io/gorm"

	"github.com/multimarket-labs/event-pod-services/database"
)

type InitService struct {
	db *database.DB
}

func NewInitService(db *database.DB) *InitService {
	return &InitService{
		db: db,
	}
}

func (s *InitService) InitDatabase() error {
	log.Info("Starting database initialization...")

	// Execute all initialization operations in a transaction
	return s.db.Transaction(func(txDB *database.DB) error {
		db := txDB.GetGorm()

		if err := s.initLanguages(db); err != nil {
			return fmt.Errorf("failed to initialize languages table: %w", err)
		}

		categoryGUIDs, err := s.initCategories(db)
		if err != nil {
			return fmt.Errorf("failed to initialize categories table: %w", err)
		}

		if err := s.initCategoryLanguages(db, categoryGUIDs); err != nil {
			return fmt.Errorf("failed to initialize category languages table: %w", err)
		}

		ecosystemGUIDs, err := s.initEcosystems(db, categoryGUIDs)
		if err != nil {
			return fmt.Errorf("failed to initialize ecosystems table: %w", err)
		}

		if err := s.initEcosystemLanguages(db, ecosystemGUIDs); err != nil {
			return fmt.Errorf("failed to initialize ecosystem languages table: %w", err)
		}

		log.Info("Database initialization completed")
		return nil
	})
}

func (s *InitService) initLanguages(db *gorm.DB) error {
	languages := []struct {
		languageName  string
		languageLabel string
		isDefault     bool
	}{
		{"en", "English", true}, // English is set as default language
		{"zh", "Chinese", false},
	}

	for _, lang := range languages {
		var existing database.Languages
		err := db.Where("language_name = ?", lang.languageName).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			newLang := database.Languages{
				LanguageName:  lang.languageName,
				LanguageLabel: lang.languageLabel,
				IsDefault:     lang.isDefault,
				IsActive:      true,
			}
			if err := db.Create(&newLang).Error; err != nil {
				return fmt.Errorf("failed to create language %s: %w", lang.languageName, err)
			}
			log.Info("Created language", "name", lang.languageName, "label", lang.languageLabel)
		} else if err != nil {
			return fmt.Errorf("failed to query language %s: %w", lang.languageName, err)
		} else {
			existing.LanguageLabel = lang.languageLabel
			existing.IsDefault = lang.isDefault
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return fmt.Errorf("failed to update language %s: %w", lang.languageName, err)
			}
			log.Info("Updated language", "name", lang.languageName, "label", lang.languageLabel)
		}
	}

	return nil
}

func (s *InitService) initCategories(db *gorm.DB) (map[string]string, error) {
	categories := []struct {
		code      string
		name      string
		sortOrder int
	}{
		{"SPORT", "Sports", 1},
		{"CRYPTO", "Cryptocurrency", 2},
		{"STOCK", "Stock", 3},
	}

	categoryGUIDs := make(map[string]string)

	for _, cat := range categories {
		existing, err := s.db.Category.GetCategoryByCode(cat.code)

		if err != nil {
			return nil, fmt.Errorf("failed to query category %s: %w", cat.code, err)
		}

		if existing == nil {
			newCategory := &database.Category{
				Code:      cat.code,
				SortOrder: cat.sortOrder,
				IsActive:  true,
				Remark:    fmt.Sprintf("Category: %s", cat.name),
			}
			if err := s.db.Category.CreateCategory(newCategory); err != nil {
				return nil, fmt.Errorf("failed to create category %s: %w", cat.code, err)
			}
			created, err := s.db.Category.GetCategoryByCode(cat.code)
			if err != nil || created == nil {
				return nil, fmt.Errorf("failed to get category %s GUID: %w", cat.code, err)
			}
			categoryGUIDs[cat.code] = created.GUID
			log.Info("Created category", "code", cat.code, "name", cat.name, "guid", created.GUID)
		} else {
			existing.SortOrder = cat.sortOrder
			existing.IsActive = true
			existing.Remark = fmt.Sprintf("Category: %s", cat.name)
			if err := s.db.Category.UpdateCategory(existing); err != nil {
				return nil, fmt.Errorf("failed to update category %s: %w", cat.code, err)
			}
			categoryGUIDs[cat.code] = existing.GUID
			log.Info("Updated category", "code", cat.code, "name", cat.name, "guid", existing.GUID)
		}
	}

	return categoryGUIDs, nil
}

func (s *InitService) initCategoryLanguages(db *gorm.DB, categoryGUIDs map[string]string) error {
	var enLang, zhLang database.Languages
	if err := db.Where("language_name = ?", "en").First(&enLang).Error; err != nil {
		return fmt.Errorf("failed to get English language GUID: %w", err)
	}
	if err := db.Where("language_name = ?", "zh").First(&zhLang).Error; err != nil {
		return fmt.Errorf("failed to get Chinese language GUID: %w", err)
	}

	categoryLanguages := map[string]map[string]struct {
		name        string
		description string
	}{
		"SPORT": {
			"en": {"Sports", "Sports events including basketball, football, etc."},
			"zh": {"运动", "运动类事件，包括篮球、足球等"},
		},
		"CRYPTO": {
			"en": {"Cryptocurrency", "Cryptocurrency market events"},
			"zh": {"加密货币", "加密货币市场事件"},
		},
		"STOCK": {
			"en": {"Stock", "Stock market events"},
			"zh": {"股票", "股票市场事件"},
		},
	}

	for code, categoryGUID := range categoryGUIDs {
		langData, ok := categoryLanguages[code]
		if !ok {
			continue
		}

		for langCode, data := range langData {
			var langGUID string
			if langCode == "en" {
				langGUID = enLang.GUID
			} else {
				langGUID = zhLang.GUID
			}

			existing, err := s.db.Category.GetCategoryLanguage(categoryGUID, langGUID)
			if err != nil {
				return fmt.Errorf("failed to query category language %s-%s: %w", code, langCode, err)
			}

			if existing == nil {
				newCategoryLang := &database.CategoryLanguage{
					CategoryGUID:       categoryGUID,
					LanguageGUID:       langGUID,
					ParentCategoryGUID: categoryGUID,
					Level:              0,
					Name:               data.name,
					Description:        data.description,
				}
				if err := s.db.Category.CreateCategoryLanguage(newCategoryLang); err != nil {
					return fmt.Errorf("failed to create category language %s-%s: %w", code, langCode, err)
				}
				log.Info("Created category language", "category", code, "language", langCode, "name", data.name)
			} else {
				existing.Name = data.name
				existing.Description = data.description
				if err := s.db.Category.UpdateCategoryLanguage(existing); err != nil {
					return fmt.Errorf("failed to update category language %s-%s: %w", code, langCode, err)
				}
				log.Info("Updated category language", "category", code, "language", langCode, "name", data.name)
			}
		}
	}

	return nil
}

func (s *InitService) initEcosystems(db *gorm.DB, categoryGUIDs map[string]string) (map[string]string, error) {
	sportCategoryGUID, ok := categoryGUIDs["SPORT"]
	if !ok {
		return nil, fmt.Errorf("sport category GUID not found")
	}

	cryptoCategoryGUID, ok := categoryGUIDs["CRYPTO"]
	if !ok {
		return nil, fmt.Errorf("crypto category GUID not found")
	}

	ecosystems := []struct {
		code         string
		name         string
		categoryCode string
		categoryGUID string
		sortOrder    int
	}{
		{"NBA", "NBA", "SPORT", sportCategoryGUID, 1},
		{"CBA", "CBA", "SPORT", sportCategoryGUID, 2},
		{"BINANCE", "Binance", "CRYPTO", cryptoCategoryGUID, 1},
		{"BYBIT", "Bybit", "CRYPTO", cryptoCategoryGUID, 2},
	}

	ecosystemGUIDs := make(map[string]string)

	for _, eco := range ecosystems {
		existing, err := s.db.Ecosystem.GetEcosystemByCode(eco.code)
		if err != nil {
			return nil, fmt.Errorf("failed to query ecosystem %s: %w", eco.code, err)
		}

		if existing == nil {
			newEcosystem := &database.Ecosystem{
				CategoryGUID: eco.categoryGUID,
				EventNum:     "1",
				Code:         eco.code,
				SortOrder:    eco.sortOrder,
				IsActive:     true,
				Remark:       fmt.Sprintf("Ecosystem: %s", eco.name),
				Extra:        database.JSONB{},
			}
			if err := s.db.Ecosystem.CreateEcosystem(newEcosystem); err != nil {
				return nil, fmt.Errorf("failed to create ecosystem %s: %w", eco.code, err)
			}
			created, err := s.db.Ecosystem.GetEcosystemByCode(eco.code)
			if err != nil || created == nil {
				return nil, fmt.Errorf("failed to get ecosystem %s GUID: %w", eco.code, err)
			}
			ecosystemGUIDs[eco.code] = created.GUID
			log.Info("Created ecosystem", "code", eco.code, "name", eco.name, "guid", created.GUID)
		} else {
			existing.CategoryGUID = eco.categoryGUID
			existing.SortOrder = eco.sortOrder
			existing.IsActive = true
			existing.Remark = fmt.Sprintf("Ecosystem: %s", eco.name)
			if err := s.db.Ecosystem.UpdateEcosystem(existing); err != nil {
				return nil, fmt.Errorf("failed to update ecosystem %s: %w", eco.code, err)
			}
			ecosystemGUIDs[eco.code] = existing.GUID
			log.Info("Updated ecosystem", "code", eco.code, "name", eco.name, "guid", existing.GUID)
		}
	}

	return ecosystemGUIDs, nil
}

func (s *InitService) initEcosystemLanguages(db *gorm.DB, ecosystemGUIDs map[string]string) error {
	var enLang, zhLang database.Languages
	if err := db.Where("language_name = ?", "en").First(&enLang).Error; err != nil {
		return fmt.Errorf("failed to get English language GUID: %w", err)
	}
	if err := db.Where("language_name = ?", "zh").First(&zhLang).Error; err != nil {
		return fmt.Errorf("failed to get Chinese language GUID: %w", err)
	}

	ecosystemLanguages := map[string]map[string]struct {
		name        string
		description string
	}{
		"NBA": {
			"en": {"NBA", "National Basketball Association"},
			"zh": {"NBA", "美国职业篮球联赛"},
		},
		"CBA": {
			"en": {"CBA", "Chinese Basketball Association"},
			"zh": {"CBA", "中国男子篮球职业联赛"},
		},
		"BINANCE": {
			"en": {"Binance", "Binance Cryptocurrency Exchange"},
			"zh": {"币安", "币安加密货币交易所"},
		},
	}

	for code, ecosystemGUID := range ecosystemGUIDs {
		langData, ok := ecosystemLanguages[code]
		if !ok {
			continue
		}

		for langCode, data := range langData {
			var langGUID string
			if langCode == "en" {
				langGUID = enLang.GUID
			} else {
				langGUID = zhLang.GUID
			}

			existing, err := s.db.Ecosystem.GetEcosystemLanguage(ecosystemGUID, langGUID)
			if err != nil {
				return fmt.Errorf("failed to query ecosystem language %s-%s: %w", code, langCode, err)
			}

			if existing == nil {
				newEcosystemLang := &database.EcosystemLanguage{
					EcosystemGUID: ecosystemGUID,
					LanguageGUID:  langGUID,
					Name:          data.name,
					Description:   data.description,
				}
				if err := s.db.Ecosystem.CreateEcosystemLanguage(newEcosystemLang); err != nil {
					return fmt.Errorf("failed to create ecosystem language %s-%s: %w", code, langCode, err)
				}
				log.Info("Created ecosystem language", "ecosystem", code, "language", langCode, "name", data.name)
			} else {
				existing.Name = data.name
				existing.Description = data.description
				if err := s.db.Ecosystem.UpdateEcosystemLanguage(existing); err != nil {
					return fmt.Errorf("failed to update ecosystem language %s-%s: %w", code, langCode, err)
				}
				log.Info("Updated ecosystem language", "ecosystem", code, "language", langCode, "name", data.name)
			}
		}
	}

	return nil
}
