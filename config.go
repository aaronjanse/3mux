package main

// FIXME: we should panic less often!

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/xdg"
	"github.com/aaronjanse/3mux/wm"
)

type UserConfig struct {
	EnableHelpBar bool                              `toml:"enable-help-bar"`
	Keys          map[string][]string               `toml:"keys"`
	Modes         map[string]map[string]interface{} `toml:"modes"`
}

type CompiledConfig struct {
	modeStarters map[string]string // key -> mode name
	isSticky     map[string]bool

	normalBindings map[string]func(*wm.Universe)
	modeBindings   map[string]map[string]func(*wm.Universe)

	generalSettings CompiledConfigGeneral
}

type CompiledConfigGeneral struct {
	enableHelpBar bool
}

func loadOrGenerateConfig() CompiledConfig {
	var userTOML string

	xdgConfigPath, err := xdg.Paths{XDGSuffix: "3mux"}.ConfigFile("config.toml")
	if err != nil {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		dirPath := filepath.Join(usr.HomeDir, ".config", "3mux")
		os.MkdirAll(dirPath, os.ModePerm)

		configPath := filepath.Join(dirPath, "config.toml")
		if _, err := os.Stat(configPath); err != nil {
			if os.IsNotExist(err) {
				userTOML = defaultConfig
				ioutil.WriteFile(configPath, []byte(defaultConfig), 0664)
			} else {
				panic("Cannot read file: " + configPath + "\n" + err.Error())
			}
		} else {
			panic(fmt.Errorf("Found in home but not in XDG? %s", os.Getenv("XDG_CONFIG_DIRS")))
		}
	} else {
		data, err := ioutil.ReadFile(xdgConfigPath)
		if err != nil {
			panic(err)
		}
		userTOML = string(data)
	}

	var conf UserConfig
	if _, err := toml.Decode(userTOML, &conf); err != nil {
		panic(err)
	}

	return compileConfig(conf)
}

func compileConfig(user UserConfig) CompiledConfig {
	conf := CompiledConfig{
		modeStarters:   map[string]string{},
		isSticky:       map[string]bool{},
		normalBindings: map[string]func(*wm.Universe){},
		modeBindings:   map[string]map[string]func(*wm.Universe){},
	}
	for modeName, mode := range user.Modes {
		sticky, ok := mode["mode-sticky"]
		if ok {
			delete(mode, "mode-sticky")
		} else {
			sticky = false
		}
		conf.isSticky[modeName] = sticky.(bool)

		if starters, ok := mode["mode-start"]; ok {
			switch x := starters.(type) {
			case []interface{}:
				for _, starter := range x {
					starter := strings.ToLower(starter.(string))
					conf.modeStarters[starter] = modeName
				}
			default:
				panic(fmt.Errorf("Expected []string: %+v (%s)", x, reflect.TypeOf(x)))
			}
			delete(mode, "mode-start")
		} else {
			panic(errors.New("Could not find starter for mode " + modeName))
		}

		mode := castMapInterface(mode)
		conf.modeBindings[modeName] = compileBindings(mode)
	}

	conf.normalBindings = compileBindings(user.Keys)

	conf.generalSettings.enableHelpBar = user.EnableHelpBar

	return conf
}

func castMapInterface(source map[string]interface{}) map[string][]string {
	out := map[string][]string{}
	for k, v := range source {
		switch x := v.(type) {
		case []interface{}:
			tmp := []string{}
			for _, abc := range x {
				tmp = append(tmp, abc.(string))
			}
			out[k] = tmp
		default:
			log.Println("Could not cast config", k, v)
		}
	}
	return out
}

func compileBindings(sourceBindings map[string][]string) map[string]func(*wm.Universe) {
	compiledBindings := map[string]func(*wm.Universe){}
	for funcName, keyCodes := range sourceBindings {
		fn, ok := configFuncBindings[funcName]
		if !ok {
			panic(errors.New("Incorrect keybinding: " + funcName))
		}
		for _, keyCode := range keyCodes {
			compiledBindings[strings.ToLower(keyCode)] = fn
		}
	}

	return compiledBindings
}

var mode = ""

func seiveConfigEvents(config CompiledConfig, u *wm.Universe, human string) bool {
	hu := strings.ToLower(human)
	if mode == "" {
		for key, theMode := range config.modeStarters {
			if hu == key {
				mode = theMode
				return true
			}
		}

		if fn, ok := config.normalBindings[hu]; ok {
			fn(u)
			return true
		}
	} else {
		bindings := config.modeBindings[mode]

		if !config.isSticky[mode] {
			mode = ""
		}

		if fn, ok := bindings[hu]; ok {
			fn(u)
			return true
		}

		mode = ""
	}
	return false
}

var configFuncBindings = map[string]func(*wm.Universe){
	"new-pane":  func(u *wm.Universe) { u.AddPane() },
	"kill-pane": func(u *wm.Universe) { u.KillPane() },

	"split-pane-horiz": func(u *wm.Universe) { u.AddPaneTmux(false) },
	"split-pane-vert":  func(u *wm.Universe) { u.AddPaneTmux(true) },

	"show-help":     func(u *wm.Universe) {},
	"hide-help-bar": func(u *wm.Universe) {},

	"toggle-fullscreen": func(u *wm.Universe) { u.ToggleFullscreen() },
	"toggle-search":     func(u *wm.Universe) { u.ToggleSearch() },

	"resize-up":    func(u *wm.Universe) { u.ResizePane(wm.Up) },
	"resize-down":  func(u *wm.Universe) { u.ResizePane(wm.Down) },
	"resize-left":  func(u *wm.Universe) { u.ResizePane(wm.Left) },
	"resize-right": func(u *wm.Universe) { u.ResizePane(wm.Right) },

	"move-pane-up":    func(u *wm.Universe) { u.MoveWindow(wm.Up) },
	"move-pane-down":  func(u *wm.Universe) { u.MoveWindow(wm.Down) },
	"move-pane-left":  func(u *wm.Universe) { u.MoveWindow(wm.Left) },
	"move-pane-right": func(u *wm.Universe) { u.MoveWindow(wm.Right) },

	"move-selection-up":    func(u *wm.Universe) { u.MoveSelection(wm.Up) },
	"move-selection-down":  func(u *wm.Universe) { u.MoveSelection(wm.Down) },
	"move-selection-left":  func(u *wm.Universe) { u.MoveSelection(wm.Left) },
	"move-selection-right": func(u *wm.Universe) { u.MoveSelection(wm.Right) },
}

const defaultConfig = `
enable-help-bar = true

[keys]

new-pane  = ['Alt+N', 'Alt+Enter']
kill-pane = ['Alt+Shift+Q']

toggle-fullscreen = ['Alt+Shift+F']
toggle-search = ['Alt+/']

show-help     = ['Alt+Shift+?']
hide-help-bar = ['Alt+\']

move-pane-up    = ['Alt+Shift+Up',    'Alt+Shift+K']
move-pane-down  = ['Alt+Shift+Down',  'Alt+Shift+J']
move-pane-left  = ['Alt+Shift+Left',  'Alt+Shift+H']
move-pane-right = ['Alt+Shift+Right', 'Alt+Shift+L']

move-selection-up    = ['Alt+Up',    'Alt+K']
move-selection-down  = ['Alt+Down',  'Alt+J']
move-selection-left  = ['Alt+Left',  'Alt+H']
move-selection-right = ['Alt+Right', 'Alt+L']

# NAME has no meaning apart from what may be displayed in a status bar
# [keys.modes.NAME]
# mode-start = ['KEYCODE'] # type KEYCODE to start this mode
# mode-sticky = STICKY
# # if STICKY:
# #   stay in this mode until we see an unrecognized key
# # else:
# #   exit this mode after the first keypress
#
# # do ACTION when we see KEYCODE while in this mode
# ACTION = ['KEYCODE']
# ACTION = ['KEYCODE']
# ACTION = ['KEYCODE']
# ...

[modes.resize]
mode-start  = ['Alt+R']
mode-sticky = true

resize-up    = ['Up',    'j']
resize-down  = ['Down',  'k']
resize-left  = ['Left',  'h']
resize-right = ['Right', 'l']

[modes.tmux]
mode-start  = ['Ctrl+B']
mode-sticky = false

split-pane-vert  = ['%']
split-pane-horiz = ['"']
move-pane-left   = ['{']
move-pane-right  = ['}']
`
