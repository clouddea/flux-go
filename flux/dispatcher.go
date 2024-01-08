package flux

type Dispatcher interface {
	Dispatch(action Action)
}

type Flux struct {
	actionCreator ActionCreator
	stores        []*Store
}

func NewFlux(actionCreator ActionCreator, stores ...*Store) *Flux {
	flux := &Flux{
		actionCreator: actionCreator,
		stores:        stores,
	}
	flux.actionCreator.setFlux(flux)
	return flux
}

func (this *Flux) Actions() ActionCreator {
	return this.actionCreator
}

func (this *Flux) Dispatch(action Action) {
	for _, store := range this.stores {
		var handler = store.Handlers[action.Name]
		if handler != nil {
			handler(store, action)
		}
		for _, controller := range store.Controllers {
			controller(store.Data)
		}
	}
}
