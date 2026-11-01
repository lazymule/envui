<p align="center"><img src="docs/icon-1024.png" width="128" alt="envui simgesi"></p>

# envui

**Türkçe** | [English](README.en.md)

Go ve Fyne ile yazılmış, `.env` dosyaları için görsel bir düzenleyici.

![Go Sürümü](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Lisans](https://img.shields.io/badge/Lisans-MIT-blue.svg)

## Neden?

`.env` dosyalarını yönetmek can sıkıcı:

- `.env.example` dosyasına göre hangi değişkenler eksik?
- Hangisi dosyada tanımlı, hangisi sistemden geliyor?
- Yanlışlıkla bir şifreyi commit ettim mi?
- Yerel ortam ile staging arasında ne fark var?

envui, birden fazla `.env` dosyasındaki ortam değişkenlerini tek bir pencereden görüp düzenlemenizi sağlar.

| Koyu tema | Açık tema |
| --- | --- |
| ![envui koyu tema](docs/screenshot-tr.png) | ![envui açık tema](docs/screenshot-tr-light.png) |

## Özellikler

- **Görsel düzenleyici**: Tüm değişkenler tabloda listelenir, düzenlemek için satıra tıklamanız yeterli.
- **Gizli değer maskeleme**: API_KEY, PASSWORD, TOKEN gibi hassas değerler otomatik algılanır ve gizlenir.
- **Çoklu dosya desteği**: `.env`, `.env.local`, `.env.staging` gibi dosyalar arasında geçiş yapabilirsiniz.
- **Eksik değişkenler**: `.env.example` dosyasında olup sizin dosyanızda olmayan değişkenler işaretlenir. Tıklayarak örnek değerle ekleyebilirsiniz.
- **Kaynak takibi**: Değişkenin dosyadan mı yoksa sistemden mi geldiği görünür.
- **Arama**: Değişkenleri ada veya değere göre filtreleyebilirsiniz.
- **Açık ve koyu tema**: GLARK kurumsal renkleriyle uyumludur. Seçiminiz bir sonraki açılışta da geçerli olur.
- **Dil desteği**: Türkçe ve İngilizce. Sistem diline göre otomatik seçilir, Ayarlar'dan değiştirilebilir.
- **Klasör hatırlama**: Klasör seçme penceresi en son açtığınız projenin yanından başlar.
- **Dosya karşılaştırma**: İki dosyayı (ör. `.env` ve `.env.staging`) yan yana görün. Farklı, eksik ve fazla değişkenler renkle işaretlenir. Bir satıra tıklayarak değeri diğer dosyaya kopyalayabilirsiniz.
- **Sorun denetimi**: İki kez tanımlanmış anahtarlar, başında/sonunda boşluk olan değerler ve `.env.example`'da dolu olup sizde boş kalanlar işaretlenir.
- **Git uyarısı**: `.env` dosyalarınız git tarafından yok sayılmıyorsa uyarır, tek tıkla `.gitignore`'a ekler.
- **Geri al / yinele**: Ctrl+Z ve Ctrl+Y ile tüm düzenlemeler geri alınabilir.
- **Diğer dosyalara kopyala**: Bir değişkeni tek seferde birden fazla dosyaya ekleyin.
- **Son projeler ve sürükle-bırak**: Son açılan projeler listelenir; klasörü pencereye sürükleyerek de açabilirsiniz.
- **Güncelleme bildirimi**: Yeni sürüm çıkınca durum çubuğunda haber verir (Ayarlar'dan kapatılabilir).
- **Yorumlar ve biçim korunur**: Kaydederken yorumlar, `export` satırları, tırnaklar ve dokunmadığınız satırlar olduğu gibi kalır.

### Klavye kısayolları

| Kısayol | İşlev |
| --- | --- |
| `Ctrl+S` | Değişen tüm dosyaları kaydet |
| `Ctrl+N` | Değişken ekle |
| `Ctrl+F` | Ara |
| `Ctrl+O` | Klasör aç |
| `Ctrl+Z` / `Ctrl+Y` | Geri al / yinele |
| `Ctrl+D` | Dosyaları karşılaştır |
| `F1` | Hakkında penceresi |

macOS'ta `Ctrl` yerine `Cmd` kullanılır.

## Kurulum

### İndirme (Windows)

[Sürümler](https://github.com/lazymule/envui/releases) sayfasından birini seçin:

| Dosya | Ne zaman? |
| --- | --- |
| `envui-setup.exe` | **Önerilen.** Kurulum sihirbazı, Başlat menüsü kısayolu ve kaldırma desteği. |
| `envui-portable.zip` | **Taşınabilir** sürüm: kurulum yok, ayarlar programın yanındaki `envui-portable.json` dosyasında tutulur. USB bellek için ideal. |
| `envui-windows-amd64.exe` | Yalnızca program dosyası (sanal makine desteği yok). |

Kurulum ve taşınabilir sürüm, sanal makinelerde ve Uzak Masaüstünde de ek bir işlem gerektirmeden çalışır (bkz. [Sorun giderme](#sorun-giderme)).

### Kaynaktan derleme

```bash
git clone https://github.com/lazymule/envui.git
cd envui
go build -o envui .
```

### Gereksinimler

- Go 1.23 veya üstü
- Bir C derleyicisi (Fyne, cgo kullanır)
- Linux'ta ek olarak şu paketler gerekebilir: `libgl1-mesa-dev xorg-dev`

### Windows için derleme

Windows'ta doğrudan derlemek için PATH'te bir C derleyicisi olmalı ([TDM-GCC](https://jmeubank.github.io/tdm-gcc/) veya MSYS2 gibi):

```bash
go build -ldflags "-H=windowsgui" -o envui.exe .
```

Linux'tan Windows için derlemek isterseniz `gcc-mingw-w64-x86-64` paketi gerekir:

```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui -s -w" -o envui.exe .
```

## Kullanım

Programı açıp bir klasör seçin:

```bash
./envui
```

Ya da bir proje klasörünü doğrudan açın:

```bash
./envui ~/Projeler/uygulamam
```

Windows'ta `envui.exe` dosyasına çift tıklamanız yeterli.

## Sorun giderme

### "WGL: The driver does not appear to support OpenGL" hatası

envui pencereyi çizmek için OpenGL 2.1 kullanır. Sanal makinelerde (VirtualBox, Hyper-V, VMware), Uzak Masaüstü (RDP) oturumlarında ve ekran kartı sürücüsü olmayan bilgisayarlarda OpenGL bulunmayabilir.

**Çözüm:** Kurulumu (`envui-setup.exe`) ya da taşınabilir sürümü (`envui-portable.zip`) kullanın. İkisinde de bir `softgl` klasörü vardır: içinde envui.exe'nin bir kopyası ve Mesa'nın yazılımsal OpenGL sürücüsü (`opengl32.dll`, `libgallium_wgl.dll`) bulunur. envui açılırken bilgisayarın OpenGL desteğini kısaca yoklar; yetersizse kendini otomatik olarak `softgl` içinden başlatır ve çizimi işlemciyle yapar. Ekran kartı olan bilgisayarlarda normal (hızlı) yol kullanılır, sistemdeki diğer programlar etkilenmez. Hakkında penceresindeki "Grafik" satırı hangi yolun kullanıldığını gösterir.

Yazılımsal çizimi zorlamak için `ENVUI_SOFTWARE_GL=1` ortam değişkenini ayarlayabilirsiniz.

Ekran kartı olan bir bilgisayara RDP ile bağlanıyorsanız, uzak bilgisayarda şu Grup İlkesi ayarını etkinleştirmek de işe yarar: *Uzak Masaüstü Oturumu Ana Bilgisayarı → Uzak Oturum Ortamı → "Tüm Uzak Masaüstü Hizmetleri oturumları için donanım grafik bağdaştırıcılarını kullan"*.

## Ayarlar

Başlıktaki ⚙ simgesiyle açılır:

- **Dil**: Türkçe veya English. Arayüz anında yenilenir, açık proje ve kaydedilmemiş değişiklikler kaybolmaz.
- **Tema**: Açık veya Koyu.
- **Açılışta**: Sistem değişkenlerinin ve gizli değerlerin gösterilip gösterilmeyeceği, güncellemelerin denetlenip denetlenmeyeceği.

Güncelleme denetimi günde en fazla bir kez GitHub'daki son sürüm bilgisini okur; başka bir veri gönderilmez.

Klasör seçme penceresi Fyne'ın kendi bileşeni olduğu için her zaman İngilizce görünür.

## Desteklenen dosyalar

envui şu dosyaları otomatik olarak bulur:

- `.env`
- `.env.local`
- `.env.development`
- `.env.staging`
- `.env.production`
- `.env.prod`
- `.env.example`
- `.env.sample`
- `.env.template`
- `.env.test`

`.env.example`, `.env.sample` ve `.env.template` şablon dosyası olarak kabul edilir. Eksik değişkenler bunlara göre belirlenir.

## Gizli değer algılama

Adında şu ifadeler geçen değişkenler otomatik olarak maskelenir:

- `PASSWORD`, `SECRET`, `TOKEN`
- `API_KEY`, `APIKEY`, `API_SECRET`
- `PRIVATE_KEY`, `ACCESS_KEY`
- `AUTH`, `CREDENTIAL`, `DSN`

## Geliştirme

```bash
# Çalıştır
go run .

# Derle
go build -o envui .

# Testleri çalıştır
go test ./...
```

Simgeyi yeniden üretmek için `python3 tools/make_icon.py` komutunu çalıştırın. Ardından Windows kaynağını güncelleyin:

```bash
x86_64-w64-mingw32-windres -O coff -i winres/envui.rc -o rsrc_windows_amd64.syso
```

Tüm Windows paketlerini (exe, kurulum, taşınabilir zip) üretmek için (`go`, `x86_64-w64-mingw32-gcc`, `7z`, `zip`, `makensis` gerekir):

```bash
tools/package_windows.sh
```

Yeni bir dil eklemek için `ui/i18n.go` dosyasındaki `messages` tablosuna bir bölüm ve `languages` listesine bir satır eklemeniz yeterli. Testler, eksik çeviri kaldıysa uyarır.

## Katkıda bulunma

Kurallar için [CONTRIBUTING.md](CONTRIBUTING.md) dosyasına bakın.

## Teşekkürler

Bu proje, **Mark Ferree**'nin [mrf/envui](https://github.com/mrf/envui) projesinden çatallanmıştır. Temeli oluşturan `.env` ayrıştırıcısı, yorum koruyan yazıcı ve ilk arayüz onun emeğidir. Projeyi açık kaynak olarak paylaştığı için kendisine çok teşekkür ederiz. 🙏

Bu çatalda arayüz yenilendi, Türkçe dil desteği, Ayarlar penceresi ve birçok küçük iyileştirme eklendi. Ayrıntılar için [CHANGELOG.md](CHANGELOG.md) dosyasına bakabilirsiniz.

## Lisans

MIT Lisansı. Ayrıntılar için [LICENSE](LICENSE) dosyasına bakın.
