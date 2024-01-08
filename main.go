package main

import (
	"bufio"
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/clouddea/flux-go/flux"
	"os"
	"strings"
)

// 定义全局有哪些Action
const (
	ACTION_TODO_C = "ACTION_TODO_C"
	ACTION_TODO_D = "ACTION_TODO_D"
)

// 定义自已的Stores
var todoList []string

var handlers = map[string]flux.Handler{
	ACTION_TODO_C: func(store *flux.Store, action flux.Action) {
		store.Data = append(store.Data.([]string), action.Payload.(string))
	},
	ACTION_TODO_D: func(store *flux.Store, action flux.Action) {
		store.Data = make([]string, 0)
	},
}

var controllers = []flux.Controller{
	func(data any) {
		line1 = fmt.Sprintf("data: %v", data)
	},
	func(data any) {
		line2 = fmt.Sprintf("len : %v", len(data.([]string)))
	},
}

// 定义一个全局的ActionCreator
type GlobalActionCreator struct {
	flux.AbstractActionCreator
}

func (this *GlobalActionCreator) CreateToDo(todo string) {
	this.GetFlux().Dispatch(flux.NewAction(ACTION_TODO_C, todo))
	this.GetFlux().Dispatch(flux.NewAction(ACTION_TODO_C, todo+" prepare"))
	this.GetFlux().Dispatch(flux.NewAction(ACTION_TODO_C, todo+" post process"))
}

func (this *GlobalActionCreator) ClearToDo() {
	this.GetFlux().Dispatch(flux.NewAction(ACTION_TODO_D, nil))
}

// 定义view
var line1 string = "data: ##########"
var line2 string = "len : ##########"

// 一个使用示例
func main() {
	tm.Clear()
	actions := &GlobalActionCreator{}
	store := flux.NewStore(todoList, handlers, controllers)
	dispatcher := flux.NewFlux(actions, store)

	// put them(dispatcher and view) together
	reader := bufio.NewReader(os.Stdin)
	for true {
		tm.MoveCursor(1, 1)
		fmt.Println(line1)
		fmt.Println(line2)
		fmt.Println("you can type cmmand to continue. ")
		fmt.Println("> add [todo] ")
		fmt.Println("> clear")
		fmt.Println("> exit")
		fmt.Print("please input : ")
		bytes, _, _ := reader.ReadLine()
		input := string(bytes)
		inputSlice := strings.Split(input, " ")
		if inputSlice[0] == "add" {
			if len(inputSlice) > 1 {
				dispatcher.Actions().(*GlobalActionCreator).CreateToDo(inputSlice[1])
			}
		} else if inputSlice[0] == "clear" {
			dispatcher.Actions().(*GlobalActionCreator).ClearToDo()
		} else if inputSlice[0] == "exit" {
			break
		}
		tm.Flush()
		tm.Clear()
	}
}
