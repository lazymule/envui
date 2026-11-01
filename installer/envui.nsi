; envui Windows installer (NSIS 3)
;
; Build (see tools/package_windows.sh):
;   makensis -DVERSION=0.3.0 -DSRCDIR=path\to\files -DOUTFILE=envui-setup.exe installer/envui.nsi
;
; SRCDIR must contain envui.exe; if it also contains opengl32.dll and
; libgallium_wgl.dll, an optional "software OpenGL" component is offered.

Unicode true
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "x64.nsh"

!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!ifndef SRCDIR
  !define SRCDIR "."
!endif
!ifndef OUTFILE
  !define OUTFILE "envui-setup.exe"
!endif

!define APPNAME   "envui"
!define PUBLISHER "Kemal BAYAT"
!define WEBSITE   "https://github.com/lazymule/envui"
!define UNINSTKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\envui"

Name "${APPNAME} ${VERSION}"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\envui"
InstallDirRegKey HKLM "Software\envui" "InstallDir"
RequestExecutionLevel admin
BrandingText "${APPNAME} ${VERSION}"

VIProductVersion "${VERSION}.0"
VIAddVersionKey /LANG=1033 "ProductName" "${APPNAME}"
VIAddVersionKey /LANG=1033 "FileDescription" "${APPNAME} installer"
VIAddVersionKey /LANG=1033 "FileVersion" "${VERSION}"
VIAddVersionKey /LANG=1033 "ProductVersion" "${VERSION}"
VIAddVersionKey /LANG=1033 "LegalCopyright" "MIT License"

!define MUI_ICON   "..\winres\envui.ico"
!define MUI_UNICON "..\winres\envui.ico"
!define MUI_ABORTWARNING
!define MUI_LANGDLL_ALLLANGUAGES
!define MUI_FINISHPAGE_RUN "$INSTDIR\envui.exe"
!define MUI_COMPONENTSPAGE_SMALLDESC

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\LICENSE"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "Turkish"
!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_RESERVEFILE_LANGDLL

LangString SecMain     ${LANG_TURKISH} "envui (gerekli)"
LangString SecMain     ${LANG_ENGLISH} "envui (required)"
LangString SecMainD    ${LANG_TURKISH} "Program dosyaları ve Başlat menüsü kısayolu."
LangString SecMainD    ${LANG_ENGLISH} "Program files and Start menu shortcut."
LangString SecDesktop  ${LANG_TURKISH} "Masaüstü kısayolu"
LangString SecDesktop  ${LANG_ENGLISH} "Desktop shortcut"
LangString SecDesktopD ${LANG_TURKISH} "Masaüstüne envui kısayolu ekler."
LangString SecDesktopD ${LANG_ENGLISH} "Adds an envui shortcut to the desktop."
LangString SecSoftGL   ${LANG_TURKISH} "Yazılımsal OpenGL (sanal makine / Uzak Masaüstü)"
LangString SecSoftGL   ${LANG_ENGLISH} "Software OpenGL (virtual machine / Remote Desktop)"
LangString SecSoftGLD  ${LANG_TURKISH} "Yalnızca 'WGL: The driver does not appear to support OpenGL' hatası alıyorsanız seçin. Mesa ile çizimi işlemci yapar."
LangString SecSoftGLD  ${LANG_ENGLISH} "Only select this if you get 'WGL: The driver does not appear to support OpenGL'. Uses Mesa to render on the CPU."

Function .onInit
  !insertmacro MUI_LANGDLL_DISPLAY
  ${IfNot} ${RunningX64}
    MessageBox MB_ICONSTOP "envui requires 64-bit Windows."
    Abort
  ${EndIf}
  SetRegView 64
FunctionEnd

Function un.onInit
  !insertmacro MUI_UNGETLANGUAGE
  SetRegView 64
FunctionEnd

Section "!$(SecMain)" SEC_MAIN
  SectionIn RO
  SetOutPath "$INSTDIR"
  File "${SRCDIR}\envui.exe"
  File "/oname=LICENSE.txt" "..\LICENSE"

  CreateShortcut "$SMPROGRAMS\envui.lnk" "$INSTDIR\envui.exe"

  WriteUninstaller "$INSTDIR\uninstall.exe"
  WriteRegStr HKLM "Software\envui" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "${UNINSTKEY}" "DisplayName" "${APPNAME}"
  WriteRegStr HKLM "${UNINSTKEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "${UNINSTKEY}" "Publisher" "${PUBLISHER}"
  WriteRegStr HKLM "${UNINSTKEY}" "URLInfoAbout" "${WEBSITE}"
  WriteRegStr HKLM "${UNINSTKEY}" "DisplayIcon" "$INSTDIR\envui.exe"
  WriteRegStr HKLM "${UNINSTKEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "${UNINSTKEY}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "${UNINSTKEY}" "NoModify" 1
  WriteRegDWORD HKLM "${UNINSTKEY}" "NoRepair" 1
SectionEnd

Section "$(SecDesktop)" SEC_DESKTOP
  CreateShortcut "$DESKTOP\envui.lnk" "$INSTDIR\envui.exe"
SectionEnd

!if /FileExists "${SRCDIR}\libgallium_wgl.dll"
Section /o "$(SecSoftGL)" SEC_SOFTGL
  SetOutPath "$INSTDIR"
  File "${SRCDIR}\opengl32.dll"
  File "${SRCDIR}\libgallium_wgl.dll"
SectionEnd
!endif

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_MAIN} $(SecMainD)
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_DESKTOP} $(SecDesktopD)
  !if /FileExists "${SRCDIR}\libgallium_wgl.dll"
    !insertmacro MUI_DESCRIPTION_TEXT ${SEC_SOFTGL} $(SecSoftGLD)
  !endif
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Section "Uninstall"
  Delete "$INSTDIR\envui.exe"
  Delete "$INSTDIR\LICENSE.txt"
  Delete "$INSTDIR\opengl32.dll"
  Delete "$INSTDIR\libgallium_wgl.dll"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"
  Delete "$SMPROGRAMS\envui.lnk"
  Delete "$DESKTOP\envui.lnk"
  DeleteRegKey HKLM "${UNINSTKEY}"
  DeleteRegKey HKLM "Software\envui"
  ; user settings (in the profile) are kept on purpose
SectionEnd
