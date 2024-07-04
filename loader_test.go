package xplugeth

import (
	"reflect"
	"testing"
)

type testInterface interface {
	Foo() string
}

type testImpl struct {}

func (*testImpl) Baz(x int) int {
	return x + 5
}

func (*testImpl) Foo() string {
	return "bar"
}

func getInitializedLoader() *pluginLoader {
	pl := &pluginLoader{
		modules: []reflect.Type{},
		hookInterfaces: []reflect.Type{},
		hooks: make(map[reflect.Type][]any),
	}
	pl.registerModule(reflect.TypeFor[testImpl]())
	pl.registerHook(reflect.TypeFor[testInterface]())
	pl.initialize()
	return pl
}

func TestLoaderBasic(t *testing.T) {
	pl := getInitializedLoader()
	ms := pl.getModules(reflect.TypeFor[testInterface]())
	if ms[0].(testInterface).Foo() != "bar" {
		t.Fatalf("unexpected value")
	}
}
func TestGetModulesByMethodName(t *testing.T) {
	pl := getInitializedLoader()
	ms := pl.getModulesByMethodName("Baz")
	if ms[0].(*testImpl).Baz(1) != 6 {
		t.Fatalf("unexpected value")
	}
}