package xplugeth

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"io/ioutil"
	"reflect"

	"github.com/go-yaml/yaml"

	"github.com/ethereum/go-ethereum/log"

)

var configPath string

type pluginLoader struct {
	modules []reflect.Type
	hookInterfaces []reflect.Type
	hooks map[reflect.Type][]any
	moduleValues []reflect.Value
	names map[string]reflect.Type
	patchsets map[reflect.Type][]Patchset
	singletons map[reflect.Type]any
	subCommands map[string]func([]string)error
}

func (pl *pluginLoader) registerHook(t reflect.Type, p ...Patchset) {
	pl.hookInterfaces = append(pl.hookInterfaces, t)
	if len(p) > 0 {
		pl.patchsets[t] = p
	}
}

func (pl *pluginLoader) registerModule(t reflect.Type, name string) {
	pl.modules = append(pl.modules, t)
	if pl.names == nil {
		n := make(map[string]reflect.Type)
		pl.names = n
	}
	pl.names[name] = t
}

func (pl *pluginLoader) registerSubCommands(provided map[string]func([]string)error) {
	for name, f := range provided {
		pl.subCommands[name] = f
	}
}

func (pl *pluginLoader) initialize(dirpath string) {
	pl.hooks = make(map[reflect.Type][]any)
	for _, mt := range pl.modules {
		mv := reflect.New(mt)
		pl.moduleValues = append(pl.moduleValues, mv)
		for _, ht := range pl.hookInterfaces {
			if reflect.PointerTo(mt).Implements(ht) {
				pl.hooks[ht] = append(pl.hooks[ht], mv.Interface())
			}
		}
	}
	configPath = dirpath
}

func (pl *pluginLoader) getModules(t reflect.Type) []any {
	return pl.hooks[t]
}

func (pl *pluginLoader) getModulesByMethodName(name string) []any {
	results := make([]any, len(pl.moduleValues))
	for i, mv := range pl.moduleValues {
		if v := mv.MethodByName(name); v.IsValid() && !v.IsZero() {
			results[i] = mv.Interface()
		}
	}
	return results
}

func (pl *pluginLoader) storeSingleton(t reflect.Type, v any) error {
	if _, ok := pl.singletons[t]; ok {
		return ErrSingletonAlreadySet
	}
	pl.singletons[t] = v
	return nil
}

func (pl *pluginLoader) getSingleton(t reflect.Type) (any, bool) {
	v, ok := pl.singletons[t]
	return v, ok
}

func (pl *pluginLoader) hasModule(name string) bool {
	_, ok := pl.names[name]
	return ok
}

func (pl *pluginLoader) hasSubcommand(commands []string) (int, bool) {
	if commands == nil || len(commands) == 0 {
		return 0, false
	}
	for i, name := range commands {
		if _, ok := pl.subCommands[name]; ok {
			log.Error("returning from has sub", "len", len(commands), "commands", commands, "name", name, "i", i)
			return i, true
		}	 
	}
	return 0, false
}

func (pl *pluginLoader) runSubcommand(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	} 
	for i, name := range args {
		command, ok := pl.subCommands[name]
		if ok {
			return true, command(args[i+1:])
		}
	}
	return false, fmt.Errorf("subcommand not %v not recognized", args)
} 

var pl *pluginLoader

func init() {
	pl = &pluginLoader{
		modules: []reflect.Type{},
		hookInterfaces: []reflect.Type{},
		hooks: make(map[reflect.Type][]any),
		singletons: make(map[reflect.Type]any),
		patchsets: make(map[reflect.Type][]Patchset),
		subCommands: make(map[string]func([]string)error),
	}
}

func RegisterModule[t any](name string) {
	pl.registerModule(reflect.TypeFor[t](), name)
}

func RegisterSubCommands(funcs map[string]func([]string)error) {
	pl.registerSubCommands(funcs)
}

func RegisterHook[t any](p ...Patchset) {
	pl.registerHook(reflect.TypeFor[t](), p...)
}

func Initialize(dirpath string) {
	pl.initialize(dirpath)
}

func GetModules[t any]() []t {
	mods := pl.getModules(reflect.TypeFor[t]())
	res := make([]t, len(mods))
	for i, m := range mods {
		res[i] = m.(t)
	}
	return res
}

func GetModulesByMethodName(name string) []any {
	return pl.getModulesByMethodName(name)
}

func StoreSingleton[t any](value t) error {
	return pl.storeSingleton(reflect.TypeFor[t](), value)
}

func GetSingleton[t any]() (t, bool) {
	v, ok := pl.getSingleton(reflect.TypeFor[t]())
	if !ok {
		var x t
		return x, ok
	}
	return v.(t), ok
}

func HasModule(name string) bool {
	return pl.hasModule(name)
}

func HasSubcommand(commands []string) (int, bool) {
	return pl.hasSubcommand(commands)
}

func RunSubcommand(commands []string) (bool, error) {
	return pl.runSubcommand(commands)
}

func GetConfig[T any](name string) (*T, bool) {

	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		log.Warn("Could not load plugins config directory, config values set to default.", "path", configPath)
		return nil, false
	}

	var fpath string
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		nameWithoutExt := strings.TrimSuffix(file.Name(), ext)
		if nameWithoutExt == name {
			if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
				log.Warn("plugin config file is not .yml or .yaml file. Skipping.", "file", file.Name())
				continue
			} else {
				fpath = path.Join(configPath, file.Name())
			}
		} else {
			log.Warn("plugin config file does not exist")
			continue
		}
	}	

	c := new(T)

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Error("error reading plugin config", "err", err)
		return nil, false
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		log.Error("error unmarshalling plugin config", "err", err)
		return nil, false
	}

	return c, true
}

func GetPatchsets() [][]Patchset {
	res := make([][]Patchset, 0, len(pl.patchsets))
	for _, patchsets := range pl.patchsets {
		res = append(res, patchsets)
	}
	return res
}