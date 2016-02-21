package main

import (
    "github.com/yuin/gopher-lua"
    "sync"
    "fmt"
)

const debug = false
const count = 10

func lstate() *lua.LState {
    l := lua.NewState()
    l.OpenLibs()
    l.PreloadModule("go", loader)
    l.SetGlobal("count", lua.LNumber(10))

    return l
}

func loader(l *lua.LState) int {
    exports := make(map[string]lua.LGFunction)

    exports["go"] = goRun
    exports["lock"] = lockNew

    mod := l.SetFuncs(l.NewTable(), exports)

    l.SetField(mod, "debug", debug && lua.LTrue || lua.LFalse)
    l.Push(mod)

    return 1
}

// Lock metatable

func checkLock(l *lua.LState) *sync.Mutex {
    ud := l.CheckUserData(1)
    if v, ok := ud.Value.(*sync.Mutex); ok {
        return v
    }
    l.ArgError(1, "sync.Mutex expected")
    return nil
}

func lockLock(l *lua.LState) int {
    lock := checkLock(l)
    lock.Lock()
    return 0
}

func lockUnlock(l *lua.LState) int {
    lock := checkLock(l)
    lock.Unlock()
    return 0
}

func lockNew(l *lua.LState) int {
    mtlock := l.NewTypeMetatable("lock")
    exports := make(map[string]lua.LGFunction)
    exports["lock"] = lockLock
    exports["unlock"] = lockUnlock
    l.SetField(mtlock, "__index", l.SetFuncs(l.NewTable(), exports))

    lock := &sync.Mutex{}
    ud := l.NewUserData()
    ud.Value = lock
    l.SetMetatable(ud, mtlock)
    l.Push(ud)

    return 1
}

// Go routine

func goRun(l *lua.LState) int {
    fn := l.ToFunction(1)

    vars := l.GetTop() - 1
    values := make([]lua.LValue, vars)
    for i := 0; i<vars; i += 1 {
        values[i] = l.Get(i+2)
    }

    go func() {
        ln := lstate()
        defer ln.Close()
        co := ln.NewThread()
        for {
            st, err, values := ln.Resume(co, fn, values...)
            if st == lua.ResumeError {
                fmt.Println("yield break(error)")
                fmt.Println(err.Error())
                break
            }

            if debug && false {
                for i, lv := range values {
                    fmt.Printf("%v : %v\n", i, lv)
                }
            }

            if st == lua.ResumeOK {
                break
            }
        }
    }()

    return 0
}

func callLua(script string, file ...bool) interface{} {
    l := lstate()
    defer l.Close()
    var err interface{}
    if len(file) == 1 && file[0] {
        err = l.DoFile(script)
    } else {
        err = l.DoString(script)
    }
    if err != nil {
        panic(err)
    }
    return l.ToInt(1)
}

func main() {
    script := "test.lua"
    res := callLua(script, true).(int)
    expect := 0
    for i := 1; i <= count; i += 1 {
        expect += i
    }
    if res != expect {
        panic(fmt.Sprintf("WTF? %d != %d", res, expect))
    }
}
