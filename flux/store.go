package flux

type ControllerSlice []Controller
type Handler func(flux Dispatcher, store *Store, action Action)

/** Store的定义 */
/** 使用Go应该抛弃使用继承*/
type Store struct {
	Name        string
	Data        any
	Handlers    map[string]Handler
	Controllers []Controller
}

func NewStore(name string, data any, handlers map[string]Handler, controllers ControllerSlice) *Store {
	return &Store{
		Name:        name,
		Data:        data,
		Handlers:    handlers,
		Controllers: controllers,
	}
}

// controller-view
type Controller func(flux Dispatcher, store *Store, data any)

// view
// view 应由用户自行定义，并由controller-view进行更新
