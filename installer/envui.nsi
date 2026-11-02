; envui Windows installer (NSIS 3)
;
; Build (see tools/package_windows.sh):
;   makensis -DVERSION=0.3.0 -DSRCDIR=path\to\files -DOUTFILE=envui-setup.exe installer/envui.nsi
;
; SRCDIR must contain envui.exe and a softgl folder (envui.exe, opengl32.dll,
; libgallium_wgl.dll). envui switches to softgl by itself on machines
; without a usable OpenGL driver.

Unicode true
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "x64.nsh"
!include "nsDialogs.nsh"
!include "LogicLib.nsh"
!include "FileFunc.nsh"

; 1 = add "Open with envui" to the context menu of .env files. Only set when
; the user ticks the box on the context-menu page (or passes /CONTEXTMENU=1
; to a silent install); never added on its own.
Var AddCtx
Var CtxCheckbox

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
; context-menu verb on all files, narrowed by AppliesTo to .env and .env.*
!define CTXKEY    "Software\Classes\*\shell\envui"
!define CTXFILTER 'System.FileName:=".env" OR System.FileName:~<".env."'

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
Page custom CtxPageShow CtxPageLeave
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
LangString CtxTitle    ${LANG_TURKISH} "Sağ tık menüsü"
LangString CtxTitle    ${LANG_ENGLISH} "Context menu"
LangString CtxSub      ${LANG_TURKISH} ".env dosyaları için $\"envui ile aç$\" seçeneği eklensin mi?"
LangString CtxSub      ${LANG_ENGLISH} "Add an $\"Open with envui$\" option for .env files?"
LangString CtxText     ${LANG_TURKISH} "İsterseniz Windows Gezgini'nde .env, .env.local, .env.example gibi dosyalara sağ tıkladığınızda $\"envui ile aç$\" seçeneği görünür.$\r$\n$\r$\nYalnızca adı .env ile başlayan dosyalarda çıkar. Varsayılan programınız değişmez; .env dosyaları çift tıklayınca yine bugünkü programla açılır. Windows 11'de seçenek $\"Daha fazla seçenek göster$\" altındadır.$\r$\n$\r$\nBu seçeneği daha sonra kurulumu yeniden çalıştırarak açıp kapatabilirsiniz."
LangString CtxText     ${LANG_ENGLISH} "If you like, an $\"Open with envui$\" option appears when you right-click .env, .env.local, .env.example and similar files in File Explorer.$\r$\n$\r$\nIt only shows for files whose name starts with .env. Your default program does not change; double-clicking a .env file still opens it as it does today. On Windows 11 the option is under $\"Show more options$\".$\r$\n$\r$\nYou can turn it on or off later by running the installer again."
LangString CtxCheck    ${LANG_TURKISH} "Evet, sağ tık menüsüne $\"envui ile aç$\" ekle"
LangString CtxCheck    ${LANG_ENGLISH} "Yes, add $\"Open with envui$\" to the context menu"
LangString MenuOpen    ${LANG_TURKISH} "envui ile aç"
LangString MenuOpen    ${LANG_ENGLISH} "Open with envui"

Function .onInit
  !insertmacro MUI_LANGDLL_DISPLAY
  ${IfNot} ${RunningX64}
    MessageBox MB_ICONSTOP "envui requires 64-bit Windows."
    Abort
  ${EndIf}
  SetRegView 64

  ; context menu: off unless the user chose it before (upgrade) or asks
  ; for it on the command line of a silent install
  StrCpy $AddCtx 0
  ReadRegStr $0 HKLM "${CTXKEY}" "MUIVerb"
  ${If} $0 != ""
    StrCpy $AddCtx 1
  ${EndIf}
  ${GetParameters} $R0
  ClearErrors
  ${GetOptions} $R0 "/CONTEXTMENU=" $R1
  ${IfNot} ${Errors}
    ${If} $R1 == "1"
      StrCpy $AddCtx 1
    ${Else}
      StrCpy $AddCtx 0
    ${EndIf}
  ${EndIf}
FunctionEnd

Function CtxPageShow
  !insertmacro MUI_HEADER_TEXT "$(CtxTitle)" "$(CtxSub)"
  nsDialogs::Create 1018
  Pop $0
  ${If} $0 == error
    Abort
  ${EndIf}
  ${NSD_CreateLabel} 0 0 100% 96u "$(CtxText)"
  Pop $0
  ${NSD_CreateCheckbox} 0 104u 100% 14u "$(CtxCheck)"
  Pop $CtxCheckbox
  ${If} $AddCtx == 1
    ${NSD_Check} $CtxCheckbox
  ${EndIf}
  nsDialogs::Show
FunctionEnd

Function CtxPageLeave
  ${NSD_GetState} $CtxCheckbox $0
  ${If} $0 == ${BST_CHECKED}
    StrCpy $AddCtx 1
  ${Else}
    StrCpy $AddCtx 0
  ${EndIf}
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

  ; software OpenGL fallback for virtual machines / Remote Desktop
  SetOutPath "$INSTDIR\softgl"
  File "${SRCDIR}\softgl\envui.exe"
  File "${SRCDIR}\softgl\opengl32.dll"
  File "${SRCDIR}\softgl\libgallium_wgl.dll"
  SetOutPath "$INSTDIR"

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

; hidden section: does what the user chose on the context-menu page
Section "-ContextMenu" SEC_CTX
  ${If} $AddCtx == 1
    WriteRegStr HKLM "${CTXKEY}" "MUIVerb" "$(MenuOpen)"
    WriteRegStr HKLM "${CTXKEY}" "Icon" '"$INSTDIR\envui.exe",0'
    WriteRegStr HKLM "${CTXKEY}" "AppliesTo" '${CTXFILTER}'
    WriteRegStr HKLM "${CTXKEY}\command" "" '"$INSTDIR\envui.exe" "%1"'
  ${Else}
    DeleteRegKey HKLM "${CTXKEY}" ; unticked on an upgrade: remove it
  ${EndIf}
  ; let Explorer pick up the change without a restart
  System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, p 0, p 0)'
SectionEnd


!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_MAIN} $(SecMainD)
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_DESKTOP} $(SecDesktopD)
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Section "Uninstall"
  Delete "$INSTDIR\envui.exe"
  Delete "$INSTDIR\LICENSE.txt"
  Delete "$INSTDIR\softgl\envui.exe"
  Delete "$INSTDIR\softgl\opengl32.dll"
  Delete "$INSTDIR\softgl\libgallium_wgl.dll"
  RMDir "$INSTDIR\softgl"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"
  Delete "$SMPROGRAMS\envui.lnk"
  Delete "$DESKTOP\envui.lnk"
  DeleteRegKey HKLM "${UNINSTKEY}"
  DeleteRegKey HKLM "${CTXKEY}"
  System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, p 0, p 0)'
  DeleteRegKey HKLM "Software\envui"
  ; user settings (in the profile) are kept on purpose
SectionEnd
