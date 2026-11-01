package env

import "strings"

// secretWords mark a variable as sensitive when they appear anywhere in its
// name. The Turkish words (sifre/şifre, parola, anahtar) match what many
// Turkish projects use, e.g. "db.parola" or "guvenlik.sifreleme_anahtari".
var secretWords = []string{
	"password", "passwd", "secret", "token", "api_key", "apikey",
	"private_key", "access_key", "auth", "credential", "dsn",
	"sifre", "parola", "anahtar",
}

// secretSegments mark a variable as sensitive when a whole part of its name
// (split at '.', '_' and '-') is one of them, e.g. WEB_SERVICE_KEY or
// db.pwd, without matching words like "keyboard".
var secretSegments = []string{"key", "pwd", "pass"}

// settingSegments mark a number/flag *about* a secret rather than the
// secret itself, e.g. guvenlik.min_parola_uzunlugu, TOKEN_EXPIRE_MINUTES or
// PASSWORD_MAX_LENGTH. They must match a whole name part; settingPrefixes
// only need to start one (Turkish suffixes: uzunlugu, uzunluk...).
var settingSegments = []string{
	"min", "max", "len", "length", "ttl", "timeout", "lifetime", "count", "limit", "enabled",
	"policy", "minutes", "seconds", "hours", "days", "dakika", "saniye", "saat", "gun", "sayisi", "kural",
}
var settingPrefixes = []string{"uzunlu", "expir"}

var turkishFold = strings.NewReplacer("ş", "s", "Ş", "s", "ı", "i", "İ", "i", "ğ", "g", "Ğ", "g",
	"ü", "u", "Ü", "u", "ö", "o", "Ö", "o", "ç", "c", "Ç", "c")

// IsSecret checks if a key name looks like it might contain a secret
func IsSecret(key string) bool {
	k := strings.ToLower(turkishFold.Replace(key))
	parts := strings.FieldsFunc(k, func(r rune) bool { return r == '.' || r == '_' || r == '-' })
	for _, part := range parts {
		for _, s := range settingSegments {
			if part == s {
				return false
			}
		}
		for _, s := range settingPrefixes {
			if strings.HasPrefix(part, s) {
				return false
			}
		}
	}
	for _, w := range secretWords {
		if strings.Contains(k, w) {
			return true
		}
	}
	for _, part := range parts {
		for _, s := range secretSegments {
			if part == s {
				return true
			}
		}
	}
	return false
}
