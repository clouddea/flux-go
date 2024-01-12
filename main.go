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

	store2 := flux.NewStore("count", 0, map[string]flux.Handler{
		ACTION_TODO_C: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			fmt.Println("update count add")
			store.Data = store.Data.(int) + 1
		},
		ACTION_TODO_D: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			fmt.Println("update count del")
			store.Data = 0
		},
	}, []flux.Controller{
		func(data any) {
			line2 = fmt.Sprintf("len : %v", data)
		},
	})

	store1 := flux.NewStore("todo", make([]string, 0),
		map[string]flux.Handler{
			ACTION_TODO_C: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
				flux.WaitFor("count")
				fmt.Println("update todo add")
				store.Data = append(store.Data.([]string), action.Payload.(string))
			},
			ACTION_TODO_D: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
				flux.WaitFor("count")
				fmt.Println("update todo del")
				store.Data = make([]string, 0)
			},
		}, []flux.Controller{
			func(data any) {
				line1 = fmt.Sprintf("data: %v", data)
			},
		})

	dispatcher := flux.NewFlux(actions, store1, store2)

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
