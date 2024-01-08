package flux

type Action struct {
	Name    string // Action 的名字
	Payload any    // Action 携带的数据
}

func NewAction(name string, payload any) Action {
	return Action{Name: name, Payload: payload}
}

type ActionCreator interface {
	setFlux(flux *Flux)
	GetFlux() *Flux
}

type AbstractActionCreator struct {
	flux *Flux
}

func (this *AbstractActionCreator) setFlux(flux *Flux) {
	if flux == nil {
		panic("bad flux!")
	}
	this.flux = flux
}

func (this *AbstractActionCreator) GetFlux() *Flux {
	return this.flux
}
