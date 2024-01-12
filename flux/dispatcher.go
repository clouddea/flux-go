package flux

/** action队列的大小 */
const FLUX_QUEUE_SIZE = 1024

type Dispatcher interface {
	Dispatch(action Action)
	WaitFor(stores ...string)
}

type Flux struct {
	/** 属性 */
	actionCreator ActionCreator
	stores        map[string]*Store
	/** 全局变量 */
	queue chan Action // 工作队列
	/** 临时变量 */
	pending Action          // 正在处理的action
	visit   map[string]bool // 正在处理的状态
}

func NewFlux(actionCreator ActionCreator, stores ...*Store) *Flux {
	storeMap := make(map[string]*Store)
	for _, store := range stores {
		if store.Name == "" {
			panic("bad store name: empty!")
		}
		if _, ok := storeMap[store.Name]; ok {
			panic("bad store name: duplicated! " + store.Name)
		}
		storeMap[store.Name] = store
	}
	flux := &Flux{
		actionCreator: actionCreator,
		stores:        storeMap,
		queue:         make(chan Action, FLUX_QUEUE_SIZE),
	}
	flux.actionCreator.setFlux(flux)
	// 开启消费者线程
	go flux.comsumer()
	return flux
}

func (this *Flux) Actions() ActionCreator {
	return this.actionCreator
}

/*
* dispatch 方法大部分情况下应该线程安全，实现方法如下
法1：使用锁
法2：使用action队列(本项目)
也可以考虑实现多线程线程安全版本，然后做实验对比
*/
func (this *Flux) Dispatch(action Action) {
	// 把action加入队列
	this.queue <- action
}

func (this *Flux) comsumer() {
	for {
		this.pending = <-this.queue
		this.visit = make(map[string]bool)
		// 依次访问每个store, 注意这里map底层是hashmap, 且遍历是随机无序的（内部取了随机数）
		for _, store := range this.stores {
			this.visitStore(store)
		}
	}
}

/**访问一个Store */
func (this *Flux) visitStore(store *Store) {
	var handler = store.Handlers[this.pending.Name]
	if handler != nil && !this.visit[store.Name] {
		this.visit[store.Name] = true
		handler(this, store, this.pending)
		for _, controller := range store.Controllers {
			controller(store.Data)
		}
	}
}

/** WaitFor实现 */
func (this *Flux) WaitFor(stores ...string) {
	for _, storeName := range stores {
		if store, ok := this.stores[storeName]; ok {
			this.visitStore(store)
		}
	}
}
