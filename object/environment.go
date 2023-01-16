package object

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outerEnv *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outerEnv
	return env
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(key string) (Object, bool) {
	obj, ok := e.store[key]

	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(key)
	}

	return obj, ok
}

func (e *Environment) Set(key string, val Object) Object {
	e.store[key] = val
	return val
}

func (e *Environment) IsKey(key string) bool {
	_, ok := e.store[key]

	if !ok && e.outer != nil {
		_, ok = e.outer.Get(key)
	}

	return ok
}
