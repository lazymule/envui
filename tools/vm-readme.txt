envui - Sanal makine / Uzak Masaüstü paketi
============================================

Bu paket, ekran kartı sürücüsü OpenGL desteklemeyen bilgisayarlar içindir:
sanal makineler (VirtualBox, Hyper-V, VMware), Uzak Masaüstü (RDP)
oturumları ve sunucular.

"WGL: The driver does not appear to support OpenGL" hatası alıyorsanız
bu paketi kullanın.

KURULUM
-------
1. Zip dosyasındaki ÜÇ dosyayı da aynı klasöre çıkarın:
     envui.exe
     opengl32.dll
     libgallium_wgl.dll
2. envui.exe dosyasına çift tıklayın.

Dosyaları ayrı klasörlere koymayın. DLL dosyaları envui.exe ile aynı
klasörde olmazsa program yine OpenGL hatası verir.

NASIL ÇALIŞIR
-------------
opengl32.dll ve libgallium_wgl.dll, Mesa projesinin yazılımsal OpenGL
sürücüsüdür (llvmpipe). Çizimi ekran kartı yerine işlemci yapar. envui
bu dosyaları kendi klasöründe bulunca otomatik olarak bu sürücüyü kullanır.
Yalnızca bu klasördeki envui etkilenir, sistemdeki diğer programlar etkilenmez.

Ekran kartı olan normal bir bilgisayarda bu pakete gerek yoktur. Orada
DLL'siz envui.exe daha hızlı çalışır.

LİSANS
------
envui: MIT Lisansı - https://github.com/lazymule/envui
Mesa @MESA_VERSION@ (opengl32.dll, libgallium_wgl.dll): MIT Lisansı
  https://www.mesa3d.org/  -  Windows derlemesi: https://github.com/pal1000/mesa-dist-win

------------------------------------------------------------------------

envui - Virtual machine / Remote Desktop package

For computers whose graphics driver has no OpenGL support (virtual
machines, Remote Desktop sessions, servers). Use it if you get
"WGL: The driver does not appear to support OpenGL".

Extract ALL THREE files (envui.exe, opengl32.dll, libgallium_wgl.dll)
into the same folder and run envui.exe. The DLLs are Mesa's software
OpenGL renderer (llvmpipe, MIT licensed); envui picks them up
automatically. Other programs on the system are not affected.
