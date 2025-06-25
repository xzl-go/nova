package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Language 语言类型
type Language string

// 支持的语言
const (
	LanguageZhCN Language = "zh-CN" // 简体中文
	LanguageEnUS Language = "en-US" // 美式英语
	LanguageJaJP Language = "ja-JP" // 日语
	LanguageKoKR Language = "ko-KR" // 韩语
)

// I18n 国际化管理器
type I18n struct {
	defaultLang Language
	messages    map[Language]map[string]string
	mu          sync.RWMutex
}

// New 创建国际化管理器
func New(defaultLang Language) *I18n {
	return &I18n{
		defaultLang: defaultLang,
		messages:    make(map[Language]map[string]string),
	}
}

// LoadMessages 加载语言包
func (i *I18n) LoadMessages(lang Language, messages map[string]string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.messages[lang] = messages
}

// LoadFromFile 从文件加载语言包
func (i *I18n) LoadFromFile(lang Language, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read language file: %v", err)
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse language file: %v", err)
	}

	i.LoadMessages(lang, messages)
	return nil
}

// LoadFromDir 从目录加载所有语言包
func (i *I18n) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read language directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		lang := Language(strings.TrimSuffix(entry.Name(), ".json"))
		filename := filepath.Join(dir, entry.Name())
		if err := i.LoadFromFile(lang, filename); err != nil {
			return fmt.Errorf("failed to load language file %s: %v", filename, err)
		}
	}

	return nil
}

// Translate 翻译消息
func (i *I18n) Translate(lang Language, key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 获取语言包
	messages, ok := i.messages[lang]
	if !ok {
		// 如果找不到指定语言，使用默认语言
		messages, ok = i.messages[i.defaultLang]
		if !ok {
			return key
		}
	}

	// 获取消息
	message, ok := messages[key]
	if !ok {
		return key
	}

	// 格式化消息
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}
	return message
}

// SetDefaultLang 设置默认语言
func (i *I18n) SetDefaultLang(lang Language) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.defaultLang = lang
}

// GetDefaultLang 获取默认语言
func (i *I18n) GetDefaultLang() Language {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.defaultLang
}

// GetSupportedLanguages 获取支持的语言列表
func (i *I18n) GetSupportedLanguages() []Language {
	i.mu.RLock()
	defer i.mu.RUnlock()

	languages := make([]Language, 0, len(i.messages))
	for lang := range i.messages {
		languages = append(languages, lang)
	}
	return languages
}

// HasLanguage 检查是否支持指定语言
func (i *I18n) HasLanguage(lang Language) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	_, ok := i.messages[lang]
	return ok
}
