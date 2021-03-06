package typex

import (
	"errors"
	"reflect"

	luajson "github.com/wwhai/gopher-json"

	"rulex/utils"

	lua "github.com/yuin/gopher-lua"
)

//
//
//
type Rule struct {
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	VM          *lua.LState `json:"-"`
	From        []string    `json:"from"`
	Actions     string      `json:"actions"`
	Success     string      `json:"success"`
	Failed      string      `json:"failed"`
}

//
// New
//
func NewRule(e RuleX,
	name string,
	description string,
	from []string,
	success string,
	actions string,
	failed string) *Rule {
	vm := lua.NewState(lua.Options{
		RegistrySize:     1024 * 20,
		RegistryMaxSize:  1024 * 80,
		RegistryGrowStep: 32,
	})
	// LoadTargetLib(e, vm)
	// LoadJqLib(e, vm)
	luajson.Preload(vm)
	return &Rule{
		Id:          utils.MakeUUID("RULE"),
		Name:        name,
		Description: description,
		From:        from,
		Actions:     actions,
		Success:     success,
		Failed:      failed,
		VM:          vm,
	}
}

// LUA Callback : Success
func (r *Rule) ExecuteSuccess() (interface{}, error) {
	return Execute(r.VM, "Success")
}

// LUA Callback : Failed

func (r *Rule) ExecuteFailed(arg lua.LValue) (interface{}, error) {
	return Execute(r.VM, "Failed", arg)
}

//
func (r *Rule) ExecuteActions(arg lua.LValue) (lua.LValue, error) {
	table := r.VM.GetGlobal("Actions")
	if table != nil && table.Type() == lua.LTTable {
		funcs := make(map[string]*lua.LFunction)
		table.(*lua.LTable).ForEach(func(idx, f lua.LValue) {
			t := reflect.TypeOf(f).Elem().Name()
			if t == "LFunction" {
				funcs[idx.String()] = f.(*lua.LFunction)
			}
		})
		return RunPipline(r.VM, funcs, arg)
	} else {
		return nil, errors.New("Actions not a lua table or not exist")
	}
}
