envui - Taşınabilir sürüm
=========================

Kurulum gerektirmez. Klasörü istediğiniz yere (USB bellek dahil) kopyalayıp
envui.exe dosyasına çift tıklayın.

Ayarlar (dil, tema, son projeler...) Windows kullanıcı profiline değil, bu
klasördeki envui-portable.json dosyasına kaydedilir. Bu dosyayı silerseniz
envui normal (kurulu) sürüm gibi davranır ve ayarları kullanıcı profiline
kaydeder.

SANAL MAKİNE / UZAK MASAÜSTÜ
----------------------------
"softgl" klasöründe Mesa'nın yazılımsal OpenGL sürücüsü vardır. envui
açılırken bilgisayarın OpenGL desteğini kontrol eder; ekran kartı yetersizse
(sanal makine, Uzak Masaüstü) çizimi otomatik olarak işlemciyle yapar.
Sizin bir şey yapmanıza gerek yoktur. Bu klasörü silmeyin.

Yazılımsal çizimi zorlamak için ENVUI_SOFTWARE_GL=1 ortam değişkenini
ayarlayabilirsiniz.

LİSANS
------
envui: MIT Lisansı - https://github.com/lazymule/envui
Mesa @MESA_VERSION@ (softgl\opengl32.dll, softgl\libgallium_wgl.dll):
  MIT Lisansı - https://www.mesa3d.org/

------------------------------------------------------------------------

envui - Portable edition

No installation needed: copy this folder anywhere (a USB stick works too)
and run envui.exe. Settings are stored in envui-portable.json in this
folder instead of your user profile; delete that file to make envui behave
like the installed version.

The "softgl" folder holds Mesa's software OpenGL renderer. envui checks the
computer's OpenGL support at start and automatically renders on the CPU
when the graphics driver isn't good enough (virtual machines, Remote
Desktop). Keep this folder. Set ENVUI_SOFTWARE_GL=1 to force it.
