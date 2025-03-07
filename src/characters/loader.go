package characters

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/carv-protocol/d.a.t.a/src/internal/conf"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/model"
)

func NewCharacter(config conf.Character, store database.Store) (*Character, error) {
	if err := store.CharacterTable().AutoMigrate(&model.Character{}); err != nil {
		return nil, err
	}

	// load from config file and compare name
	character, err := loadFromFile(config.Path)
	if err != nil {
		return nil, fmt.Errorf("load from file: %w", err)
	}
	if config.Name == "" {
		config.Name = character.Name
	}
	if config.Name != character.Name {
		return nil, fmt.Errorf("character name not match")
	}
	if config.Name == "" {
		return nil, fmt.Errorf("no character name found")
	}

	// check db. if exists, load from db
	characterDB, err := loadFromDB(store, config.Name)
	if err != nil {
		return nil, fmt.Errorf("load from db: %w", err)
	}
	if characterDB != nil {
		return characterDB, nil
	}

	// if non-exists, load from config file and write to db
	if err = writeToDB(store, character); err != nil {
		return nil, fmt.Errorf("write to db: %w", err)
	}

	return character, nil
}

func loadFromDB(store database.Store, name string) (*Character, error) {

	var characterDB model.Character
	if err := store.CharacterTable().Where("name = ?", name).Find(&characterDB).Error; err != nil {
		return nil, err
	}
	if characterDB.ID == 0 {
		return nil, nil
	}

	var (
		bio              []string
		lore             []string
		styleGuide       StyleGuide
		topics           []string
		goals            []Goal
		messageExamples  []string
		priorityAccounts []Account
		preferences      map[string]float64
	)

	if err := json.Unmarshal([]byte(characterDB.Bio), &bio); err != nil {
		return nil, fmt.Errorf("unmarshal bio err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.Lore), &lore); err != nil {
		return nil, fmt.Errorf("unmarshal lore err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.Style), &styleGuide); err != nil {
		return nil, fmt.Errorf("unmarshal styleGuide err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.Topics), &topics); err != nil {
		return nil, fmt.Errorf("unmarshal topics err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.Goals), &goals); err != nil {
		return nil, fmt.Errorf("unmarshal goals err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.MessageExamples), &messageExamples); err != nil {
		return nil, fmt.Errorf("unmarshal messageExamples err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.PriorityAccounts), &priorityAccounts); err != nil {
		return nil, fmt.Errorf("unmarshal priorityAccounts err: %w", err)
	}
	if err := json.Unmarshal([]byte(characterDB.Preferences), &preferences); err != nil {
		return nil, fmt.Errorf("unmarshal preferences err: %w", err)
	}

	return &Character{
		Name:             characterDB.Name,
		System:           characterDB.System,
		Bio:              bio,
		Lore:             lore,
		Style:            styleGuide,
		Topics:           topics,
		Goals:            goals,
		MessageExamples:  messageExamples,
		TaskInstructions: characterDB.TaskInstructions,
		PriorityAccounts: priorityAccounts,
		Preferences:      preferences,
	}, nil

}

func writeToDB(store database.Store, character *Character) error {
	bio, err := json.Marshal(character.Bio)
	if err != nil {
		return fmt.Errorf("marshal bio err: %w", err)
	}
	lore, err := json.Marshal(character.Lore)
	if err != nil {
		return fmt.Errorf("marshal lore err: %w", err)
	}
	style, err := json.Marshal(character.Style)
	if err != nil {
		return fmt.Errorf("marshal style err: %w", err)
	}
	topics, err := json.Marshal(character.Topics)
	if err != nil {
		return fmt.Errorf("marshal topics err: %w", err)
	}
	goals, err := json.Marshal(character.Goals)
	if err != nil {
		return fmt.Errorf("marshal goals err: %w", err)
	}
	messageExamples, err := json.Marshal(character.MessageExamples)
	if err != nil {
		return fmt.Errorf("marshal messageExamples err: %w", err)
	}
	priorityAccounts, err := json.Marshal(character.PriorityAccounts)
	if err != nil {
		return fmt.Errorf("marshal priorityAccounts err: %w", err)
	}
	preferences, err := json.Marshal(character.Preferences)
	if err != nil {
		return fmt.Errorf("marshal preferences err: %w", err)
	}

	return store.CharacterTable().Create(&model.Character{
		Name:             character.Name,
		System:           character.System,
		Bio:              string(bio),
		Lore:             string(lore),
		Style:            string(style),
		Topics:           string(topics),
		Goals:            string(goals),
		MessageExamples:  string(messageExamples),
		TaskInstructions: character.TaskInstructions,
		PriorityAccounts: string(priorityAccounts),
		Preferences:      string(preferences),
	}).Error
}

func loadFromFile(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var config CharacterConfig
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	return &Character{
		Name:             config.Name,
		System:           config.System,
		Bio:              config.Bio,
		Lore:             config.Lore,
		Style:            StyleGuide(config.Style),
		Topics:           config.Topics,
		Goals:            config.Goals,
		PriorityAccounts: config.PriorityAccounts,
		Preferences:      config.Preferences,
		MessageExamples:  config.MessageExamples,
		TaskInstructions: config.TaskInstructions,
	}, nil
}
