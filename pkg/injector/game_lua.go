package injector

// here, we define where in main.lua we will inject the loader start code.
// The 'right place' could be different for different game versions,
// so we need to keep a map of file MD5sum to the correct injection point.

const startCode = "initSteamodded()"

type InjectionPoint struct {
	Line        int
	Indentation int
}

var gameLuaInjectionPoints = map[string]InjectionPoint{
	"cc95a80eef5ae01641621abdb670771d": {Line: 206, Indentation: 4},
}
