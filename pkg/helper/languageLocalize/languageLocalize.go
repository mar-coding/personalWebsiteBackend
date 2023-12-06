package languageLocalize

import (
	"context"
	"embed"
	"errors"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
	"strings"
)

const _defaultLocalizeKey = "localize"

type I18n struct {
	bundle *i18n.Bundle
}

// New create I18n object to get localize and message translation
func New(languageFS embed.FS, languagePath []string, Unmarshalers map[string]func(data []byte, v interface{}) error) (*I18n, error) {
	bundle := i18n.NewBundle(language.English)

	for format, unmarshalFunc := range Unmarshalers {
		bundle.RegisterUnmarshalFunc(format, unmarshalFunc)

	}

	for _, path := range languagePath {
		_, err := bundle.LoadMessageFileFS(languageFS, path)
		if err != nil {
			return nil, errors.New("language-localize: failed to load message from fs file")
		}
	}

	return &I18n{
		bundle: bundle,
	}, nil
}

// GetLocalizeFromContext object from context
func GetLocalizeFromContext(ctx context.Context) (*i18n.Localizer, error) {
	localize, ok := ctx.Value(_defaultLocalizeKey).(*i18n.Localizer)
	if !ok {
		return nil, errors.New("language-localize: failed to get Localize object from context")
	}
	return localize, nil
}

// GetLocalize from bundle base on languages
func (i *I18n) GetLocalize(languages ...string) *i18n.Localizer {
	return i18n.NewLocalizer(i.bundle, languages...)
}

// GetLanguageFromMD get language from MD
func (i *I18n) GetLanguageFromMD(md metadata.MD) string {
	lang := "en"
	acceptLang := md.Get("accept-language")
	if len(acceptLang) != 0 {
		lang = strings.Split(acceptLang[0], ",")[0]
	} else {
		acceptLang = md.Get("X-Client-Language")
		if len(acceptLang) != 0 {
			lang = strings.Split(acceptLang[0], ",")[0]
		}
	}
	return lang
}
