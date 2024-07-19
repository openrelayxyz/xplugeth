package xplugeth

import (
	"reflect"
)

type pluginLoader struct {
	modules []reflect.Type
	hookInterfaces []reflect.Type
	hooks map[reflect.Type][]any
	moduleValues []reflect.Value

	singletons map[reflect.Type]any
}

func (pl *pluginLoader) registerHook(t reflect.Type) {
	pl.hookInterfaces = append(pl.hookInterfaces, t)
}

func (pl *pluginLoader) registerModule(t reflect.Type) {
	pl.modules = append(pl.modules, t)
}

func (pl *pluginLoader) initialize() {
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

var pl *pluginLoader

func init() {
	pl = &pluginLoader{
		modules: []reflect.Type{},
		hookInterfaces: []reflect.Type{},
		hooks: make(map[reflect.Type][]any),
		singletons: make(map[reflect.Type]any),
	}
}

func RegisterModule[t any]() {
	pl.registerModule(reflect.TypeFor[t]())
}

func RegisterHook[t any]() {
	pl.registerHook(reflect.TypeFor[t]())
}

func Initialize() {
	pl.initialize()
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