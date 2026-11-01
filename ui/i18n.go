package ui

import (
	"fmt"
	"os"
	"strings"
)

// Lang is a UI language code.
type Lang string

const (
	LangEnglish Lang = "en"
	LangTurkish Lang = "tr"
)

// languages lists the supported languages in the order shown in Settings.
var languages = []struct {
	Code Lang
	Name string
}{
	{LangTurkish, "Türkçe"},
	{LangEnglish, "English"},
}

var currentLang = LangEnglish

// SetLanguage changes the UI language. Unknown codes fall back to English.
func SetLanguage(l Lang) {
	if _, ok := messages[l]; ok {
		currentLang = l
	} else {
		currentLang = LangEnglish
	}
}

// CurrentLanguage returns the active UI language.
func CurrentLanguage() Lang { return currentLang }

// detectLanguage guesses the language from the operating system.
func detectLanguage() Lang {
	candidates := systemLanguages()
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		if s := os.Getenv(v); s != "" {
			candidates = append(candidates, s)
		}
	}
	for _, c := range candidates {
		c = strings.ToLower(c)
		for code := range messages {
			if strings.HasPrefix(c, string(code)) {
				return code
			}
		}
	}
	return LangEnglish
}

// T returns the translation for key, formatted with args.
func T(key string, args ...any) string {
	s, ok := messages[currentLang][key]
	if !ok {
		if s, ok = messages[LangEnglish][key]; !ok {
			s = key
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

// Tn returns a count-dependent translation: key+".one" when n == 1,
// key+".other" otherwise. The count is passed as the first format argument.
func Tn(key string, n int) string {
	if n == 1 {
		if _, ok := messages[currentLang][key+".one"]; ok {
			return T(key+".one", n)
		}
	}
	return T(key+".other", n)
}

var messages = map[Lang]map[string]string{
	LangEnglish: {
		"app.title": "envui",

		"header.noProject":    "No project open",
		"header.open":         "Open folder",
		"header.settings":     "Settings",
		"sidebar.files":       "FILES",
		"sidebar.newFile":     "New .env file",
		"toolbar.shell":       "Shell variables",
		"toolbar.show":        "Show secrets",
		"toolbar.hide":        "Hide secrets",
		"toolbar.add":         "Add variable",
		"toolbar.search":      "Search variables…  (Ctrl+F)",
		"status.unsaved":      "Unsaved changes",
		"status.save":         "Save  (Ctrl+S)",
		"status.saved":        "Saved %s",
		"status.created":      "Created %s",
		"status.reloaded":     "Reloaded from disk",
		"status.copied":       "Copied %s to clipboard",
		"status.missing":      "%d missing from %s",
		"status.inSync":       "in sync with %s",
		"stats.vars.one":      "%d variable",
		"stats.vars.other":    "%d variables",
		"stats.secrets.one":   "%d secret",
		"stats.secrets.other": "%d secrets",
		"stats.matches.one":   "%d match",
		"stats.matches.other": "%d matches",

		"empty.open.title":   "Open a project",
		"empty.open.message": "Choose a folder that contains .env files to view and edit its environment variables.",
		"empty.none.title":   "No .env files found",
		"empty.none.message": "%s does not contain any .env files yet.",
		"empty.none.button":  "Create .env",

		"col.key":    "VARIABLE",
		"col.value":  "VALUE",
		"col.source": "SOURCE",

		"source.file":    "file",
		"source.missing": "missing",
		"source.shell":   "shell",

		"cell.notSetExample": "not set (example: %s)",
		"cell.notSet":        "not set - click to add",
		"cell.empty":         "empty",

		"common.cancel": "Cancel",
		"common.ok":     "OK",
		"common.yes":    "Yes",
		"common.no":     "No",
		"common.error":  "Error",
		"common.save":   "Save",
		"common.add":    "Add",
		"common.delete": "Delete",
		"common.create": "Create",

		"var.addTitle":    "Add variable",
		"var.editTitle":   "Edit variable",
		"var.name":        "Name",
		"var.value":       "Value",
		"var.valueHint":   "value",
		"var.secretHint":  "Looks like a secret — the value will be masked in the table.",
		"var.exists":      "%s already exists in this file.",
		"var.deleteTitle": "Delete variable",
		"var.deleteAsk":   "Delete %s from %s?",
		"var.noFileTitle": "No file selected",
		"var.noFileText":  "Open a project and select a file first.",
		"key.required":    "Name is required",
		"key.invalid":     "Use letters, digits and _ (must not start with a digit)",

		"newFile.title":         "New env file",
		"newFile.name":          "File name",
		"newFile.copy":          "Copy variables from %s",
		"newFile.noProject":     "No project",
		"newFile.noProjectText": "Open a project folder first.",
		"newFile.invalid":       "Invalid file name %q",
		"newFile.exists":        "%s already exists",

		"err.notFolder": "%s is not a folder",
		"err.saving":    "Saving %s: %v",

		"unsaved.title":       "Unsaved changes",
		"unsaved.discardAsk":  "You have unsaved changes. Discard them?",
		"unsaved.discard":     "Discard",
		"unsaved.quitAsk":     "You have unsaved changes. Save them before quitting?",
		"unsaved.saveAndQuit": "Save and quit",

		"settings.title":    "Settings",
		"settings.language": "Language",
		"settings.theme":    "Theme",
		"settings.light":    "Light",
		"settings.dark":     "Dark",
		"settings.shell":    "Show shell variables",
		"settings.secrets":  "Show secrets",
		"settings.startup":  "On start",
		"settings.note":     "Some system dialogs (such as the folder picker) are always shown in English.",

		"common.close": "Close",

		"about.title":            "About envui",
		"about.version":          "Version %s",
		"about.description":      "A visual editor for .env files: see, compare and edit environment variables across multiple files.",
		"about.developer":        "Developer",
		"about.developerName":    "Kemal BAYAT",
		"about.original":         "Original project",
		"about.source":           "Source code",
		"about.license":          "License",
		"about.licenseText":      "MIT License",
		"about.builtWith":        "Built with",
		"about.renderer":         "Graphics",
		"about.rendererHardware": "OpenGL (graphics card)",
		"about.rendererSoftware": "Mesa software rendering (%s)",
		"about.system":           "System",
		"about.thanks":           "Thanks to Mark Ferree for creating envui and sharing it as open source.",

		"recent.none":         "No recent projects",
		"recent.clear":        "Clear list",
		"recent.title":        "Recent projects",
		"empty.dropHint":      "You can also drag a folder onto this window.",
		"issues.count.one":    "%d problem",
		"issues.count.other":  "%d problems",
		"issues.title":        "Problems",
		"issues.intro":        "Possible mistakes in %s. Click one to fix it:",
		"issue.duplicate":     "defined more than once (the last one wins)",
		"issue.spaces":        "value starts or ends with spaces",
		"issue.emptyRequired": "empty, but %s has a value for it",
		"git.warning":         "Not ignored by git: %s. Secrets in these files could be committed.",
		"git.add":             "Add to .gitignore",
		"git.title":           "Git",
		"git.added":           "Added to .gitignore",
		"git.stillTracked":    "The files were added to .gitignore, but another git rule still includes them. Please check your .gitignore files.",
		"compare.button":      "Compare",
		"compare.title":       "Compare files",
		"compare.needTwo":     "The project needs at least two env files.",
		"compare.colKey":      "VARIABLE",
		"compare.colLeft":     "LEFT",
		"compare.colRight":    "RIGHT",
		"compare.colStatus":   "STATUS",
		"compare.summary":     "%d different · %d only left · %d only right · %d same",
		"compare.same":        "same",
		"compare.changed":     "different",
		"compare.onlyLeft":    "only left",
		"compare.onlyRight":   "only right",
		"compare.onlyDiff":    "Only differences",
		"compare.hint":        "Click a row to copy the value to the other file.",
		"compare.copyTo":      "Copy to %s",
		"compare.copyAsk":     "Copy the value of %s:",
		"status.copiedTo":     "%s copied to %s",
		"status.undone":       "Undone",
		"status.redone":       "Redone",
		"copyTo.button":       "Copy to other files…",
		"copyTo.title":        "Copy to other files",
		"copyTo.intro":        "Copy %s with this value into:",
		"copyTo.missing":      "missing",
		"copyTo.same":         "same value",
		"copyTo.different":    "different value",
		"copyTo.copy":         "Copy",
		"update.title":        "Updates",
		"update.check":        "Check for updates",
		"update.failed":       "Could not check for updates: %v",
		"update.upToDate":     "You have the latest version (%s).",
		"update.available":    "envui %s is available (you have %s). Open the download page?",
		"update.open":         "Open download page",
		"update.notice":       "New version: %s",
		"settings.updates":    "Check for updates (once a day)",
		"about.portable":      "portable",
	},

	LangTurkish: {
		"app.title": "envui",

		"header.noProject":    "Açık proje yok",
		"header.open":         "Klasör aç",
		"header.settings":     "Ayarlar",
		"sidebar.files":       "DOSYALAR",
		"sidebar.newFile":     "Yeni .env dosyası",
		"toolbar.shell":       "Sistem değişkenleri",
		"toolbar.show":        "Gizlileri göster",
		"toolbar.hide":        "Gizlileri sakla",
		"toolbar.add":         "Değişken ekle",
		"toolbar.search":      "Değişken ara…  (Ctrl+F)",
		"status.unsaved":      "Kaydedilmemiş değişiklikler",
		"status.save":         "Kaydet  (Ctrl+S)",
		"status.saved":        "Kaydedildi: %s",
		"status.created":      "Oluşturuldu: %s",
		"status.reloaded":     "Diskten yeniden yüklendi",
		"status.copied":       "%s panoya kopyalandı",
		"status.missing":      "%[2]s dosyasına göre %[1]d eksik",
		"status.inSync":       "%s ile uyumlu",
		"stats.vars.other":    "%d değişken",
		"stats.secrets.other": "%d gizli",
		"stats.matches.other": "%d sonuç",

		"empty.open.title":   "Bir proje açın",
		"empty.open.message": "Ortam değişkenlerini görüntülemek ve düzenlemek için .env dosyaları içeren bir klasör seçin.",
		"empty.none.title":   ".env dosyası bulunamadı",
		"empty.none.message": "%s klasöründe henüz .env dosyası yok.",
		"empty.none.button":  ".env oluştur",

		"col.key":    "DEĞİŞKEN",
		"col.value":  "DEĞER",
		"col.source": "KAYNAK",

		"source.file":    "dosya",
		"source.missing": "eksik",
		"source.shell":   "sistem",

		"cell.notSetExample": "tanımlı değil (örnek: %s)",
		"cell.notSet":        "tanımlı değil - eklemek için tıklayın",
		"cell.empty":         "boş",

		"common.cancel": "İptal",
		"common.ok":     "Tamam",
		"common.yes":    "Evet",
		"common.no":     "Hayır",
		"common.error":  "Hata",
		"common.save":   "Kaydet",
		"common.add":    "Ekle",
		"common.delete": "Sil",
		"common.create": "Oluştur",

		"var.addTitle":    "Değişken ekle",
		"var.editTitle":   "Değişkeni düzenle",
		"var.name":        "Ad",
		"var.value":       "Değer",
		"var.valueHint":   "değer",
		"var.secretHint":  "Gizli bir değer gibi görünüyor — tabloda maskelenerek gösterilecek.",
		"var.exists":      "%s bu dosyada zaten var.",
		"var.deleteTitle": "Değişkeni sil",
		"var.deleteAsk":   "%s, %s dosyasından silinsin mi?",
		"var.noFileTitle": "Dosya seçilmedi",
		"var.noFileText":  "Önce bir proje açıp bir dosya seçin.",
		"key.required":    "Ad gerekli",
		"key.invalid":     "Harf, rakam ve _ kullanın (rakamla başlayamaz, Türkçe karakter olmaz)",

		"newFile.title":         "Yeni env dosyası",
		"newFile.name":          "Dosya adı",
		"newFile.copy":          "Değişkenleri %s dosyasından kopyala",
		"newFile.noProject":     "Proje yok",
		"newFile.noProjectText": "Önce bir proje klasörü açın.",
		"newFile.invalid":       "Geçersiz dosya adı: %q",
		"newFile.exists":        "%s zaten var",

		"err.notFolder": "%s bir klasör değil",
		"err.saving":    "%s kaydedilemedi: %v",

		"unsaved.title":       "Kaydedilmemiş değişiklikler",
		"unsaved.discardAsk":  "Kaydedilmemiş değişiklikleriniz var. Vazgeçilsin mi?",
		"unsaved.discard":     "Kaydetme",
		"unsaved.quitAsk":     "Kaydedilmemiş değişiklikleriniz var. Çıkmadan önce kaydedilsin mi?",
		"unsaved.saveAndQuit": "Kaydet ve çık",

		"settings.title":    "Ayarlar",
		"settings.language": "Dil",
		"settings.theme":    "Tema",
		"settings.light":    "Açık",
		"settings.dark":     "Koyu",
		"settings.shell":    "Sistem değişkenlerini göster",
		"settings.secrets":  "Gizli değerleri göster",
		"settings.startup":  "Açılışta",
		"settings.note":     "Bazı sistem pencereleri (ör. klasör seçici) her zaman İngilizce görünür.",

		"common.close": "Kapat",

		"about.title":            "envui Hakkında",
		"about.version":          "Sürüm %s",
		"about.description":      ".env dosyaları için görsel düzenleyici: birden fazla dosyadaki ortam değişkenlerini görün, karşılaştırın ve düzenleyin.",
		"about.developer":        "Geliştiren",
		"about.developerName":    "Kemal BAYAT",
		"about.original":         "Orijinal proje",
		"about.source":           "Kaynak kod",
		"about.license":          "Lisans",
		"about.licenseText":      "MIT Lisansı",
		"about.builtWith":        "Kullanılan teknoloji",
		"about.renderer":         "Grafik",
		"about.rendererHardware": "OpenGL (ekran kartı)",
		"about.rendererSoftware": "Mesa yazılımsal çizim (%s)",
		"about.system":           "Sistem",
		"about.thanks":           "envui'yi geliştirip açık kaynak olarak paylaştığı için Mark Ferree'ye teşekkürler.",

		"recent.none":         "Son proje yok",
		"recent.clear":        "Listeyi temizle",
		"recent.title":        "Son projeler",
		"empty.dropHint":      "Bir klasörü bu pencereye sürükleyip bırakarak da açabilirsiniz.",
		"issues.count.other":  "%d sorun",
		"issues.title":        "Sorunlar",
		"issues.intro":        "%s dosyasındaki olası hatalar. Düzeltmek için birine tıklayın:",
		"issue.duplicate":     "birden fazla kez tanımlı (sonuncusu geçerli olur)",
		"issue.spaces":        "değer boşlukla başlıyor ya da bitiyor",
		"issue.emptyRequired": "boş, ama %s dosyasında bir değeri var",
		"git.warning":         "Git tarafından yok sayılmıyor: %s. Bu dosyalardaki gizli bilgiler commit'lenebilir.",
		"git.add":             ".gitignore'a ekle",
		"git.title":           "Git",
		"git.added":           ".gitignore'a eklendi",
		"git.stillTracked":    "Dosyalar .gitignore'a eklendi ama başka bir git kuralı onları hâlâ dahil ediyor. Lütfen .gitignore dosyalarınızı kontrol edin.",
		"compare.button":      "Karşılaştır",
		"compare.title":       "Dosyaları karşılaştır",
		"compare.needTwo":     "Projede en az iki env dosyası olmalı.",
		"compare.colKey":      "DEĞİŞKEN",
		"compare.colLeft":     "SOL",
		"compare.colRight":    "SAĞ",
		"compare.colStatus":   "DURUM",
		"compare.summary":     "%d farklı · %d yalnız solda · %d yalnız sağda · %d aynı",
		"compare.same":        "aynı",
		"compare.changed":     "farklı",
		"compare.onlyLeft":    "yalnız solda",
		"compare.onlyRight":   "yalnız sağda",
		"compare.onlyDiff":    "Yalnızca farklar",
		"compare.hint":        "Değeri diğer dosyaya kopyalamak için bir satıra tıklayın.",
		"compare.copyTo":      "%s dosyasına kopyala",
		"compare.copyAsk":     "%s değerini kopyala:",
		"status.copiedTo":     "%s → %s kopyalandı",
		"status.undone":       "Geri alındı",
		"status.redone":       "Yinelendi",
		"copyTo.button":       "Diğer dosyalara kopyala…",
		"copyTo.title":        "Diğer dosyalara kopyala",
		"copyTo.intro":        "%s bu değerle şu dosyalara kopyalansın:",
		"copyTo.missing":      "yok",
		"copyTo.same":         "aynı değer",
		"copyTo.different":    "farklı değer",
		"copyTo.copy":         "Kopyala",
		"update.title":        "Güncellemeler",
		"update.check":        "Güncellemeleri denetle",
		"update.failed":       "Güncellemeler denetlenemedi: %v",
		"update.upToDate":     "En son sürümü kullanıyorsunuz (%s).",
		"update.available":    "envui %s yayımlandı (sizdeki: %s). İndirme sayfası açılsın mı?",
		"update.open":         "İndirme sayfasını aç",
		"update.notice":       "Yeni sürüm: %s",
		"settings.updates":    "Güncellemeleri denetle (günde bir kez)",
		"about.portable":      "taşınabilir",
	},
}
