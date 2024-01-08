package flux

type ControllerSlice []Controller
type Handler func(store *Store, action Action)

/** Store的定义 */
/** 使用Go应该抛弃使用继承*/
type Store struct {
	Data        any
	Handlers    map[string]Handler
	Controllers []Controller
}

func NewStore(data any, handlers map[string]Handler, controllers ControllerSlice) *Store {
	return &Store{
		Data:        data,
		Handlers:    handlers,
		Controllers: controllers,
	}
}

// controller-view
type Controller func(data any)

// view
// view 应由用户自行定义，并由controller-view进行更新
