package gopherlua

import lua "github.com/yuin/gopher-lua"

func GoMapToLuaTable(L *lua.LState, m map[string]interface{}) *lua.LTable {
	table := L.NewTable()
	for key, value := range m {
		var luaValue lua.LValue

		// 根据值的类型选择合适的 Lua 类型
		switch v := value.(type) {
		case string:
			luaValue = lua.LString(v)
		case int, int32, int64, uint, uint32, uint64, float32, float64:
			tv := v.(float64)
			luaValue = lua.LNumber(tv)
		case bool:
			luaValue = lua.LBool(v)
		case map[string]interface{}:
			luaValue = GoMapToLuaTable(L, v) // 递归处理嵌套 map
		default:
			luaValue = lua.LNil // 不支持的类型
		}

		table.RawSetString(key, luaValue)
	}
	return table
}
