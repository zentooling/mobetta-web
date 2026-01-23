package infra

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type LangService struct {
	bundle    *i18n.Bundle
	ctx       *gin.Context
	localizer *i18n.Localizer
}

func NewLangService(ctx *gin.Context, bundle *i18n.Bundle) LangService {
	localizer := i18n.NewLocalizer(bundle, ctx.Request.Header.Get("Accept-Language"), "en")
	return LangService{
		bundle:    bundle,
		ctx:       ctx,
		localizer: localizer,
	}
}

func (s *LangService) Trans(str string) string {
	// TODO, modify this to handle plural and more types of phrases
	for _, m := range translationMessages {
		if m.ID == str {
			localizedString, _ := s.localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &m,
			})
			return localizedString
		} else if m.Other == str {
			localizedString, _ := s.localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &m,
			})
			return localizedString
		}
	}
	return str
}

func LoadLanguageBundles() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	languages := []string{
		"en",
		"sv",
	}
	for _, l := range languages {
		_, err := bundle.LoadMessageFile(fmt.Sprintf("active.%s.toml", l))
		if err != nil {
			slog.Error("Run", "error", err)
			os.Exit(1)
		}
	}
	// side effect during refactoring
	return bundle
}
