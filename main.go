package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	ap := app.New()
	//icon, _ := fyne.LoadResourceFromPath("./assets/img/chrome.ico")
	ap.SetIcon(resourceAssetsImgChromeIco)
	//t.SetFonts("./assets/font/MiSans-Regular.ttf", "")
	//初始化绑定数据
	data := initData()
	initBundle(*data)
	ap.Settings().SetTheme(&MyTheme{data.themeSettings, data.langSettings})
	meta := ap.Metadata()
	win := ap.NewWindow(LoadString("TitleLabel") + " v" + meta.Version + " by Libs")
	tabs := container.NewAppTabs(
		container.NewTabItem(LoadString("TabMainLabel"), baseScreen(win, data)),
		container.NewTabItem("Chrome++", chromePlusScreen(win, data)),
		container.NewTabItem(LoadString("TabSettingLabel"), settingsScreen(ap, win, data)),
	)
	tabs.Refresh()
	tabs.OnSelected = func(t *container.TabItem) {
		fyne.CurrentApp().Settings().SetTheme(fyne.CurrentApp().Settings().Theme())
	}

	win.SetContent(
		tabs,
	)
	//win.SetMainMenu(makeMenu(ap, win))
	win.CenterOnScreen()
	win.Resize(fyne.NewSize(500, 400))
	//win.SetFixedSize(true)
	win.SetOnClosed(func() {
		//保存配置数据
		saveConfig(data)
	})
	win.ShowAndRun()
}

func initBundle(data SettingsData) {
	lang := getString(data.langSettings)
	if lang == "System" || lang == "" {
		DelayInitializeLocale()
	} else {
		SetLocale(lang)
	}
}
