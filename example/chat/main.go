package main

import (
	"bufio"
	"fmt"
	"github.com/clouddea/flux-go/flux"
	"os"
	"strconv"
	"strings"
	"time"
)

type Screen struct {
	w    int
	h    int
	data [][]byte
}

func NewScreen(w, h int) *Screen {
	screen := &Screen{
		w:    w,
		h:    h,
		data: make([][]byte, h),
	}
	for i := 0; i < h; i++ {
		screen.data[i] = make([]byte, w)
	}
	return screen
}

func (this *Screen) Clear() {
	for i := 0; i < this.h; i++ {
		for j := 0; j < this.w; j++ {
			this.data[i][j] = ' '
		}
	}
	for i := 0; i < this.w; i++ {
		this.data[0][i] = '-'
		this.data[this.h-1][i] = '-'
		this.data[2][i] = '-'
	}
	for i := 0; i < this.h; i++ {
		this.data[i][0] = '|'
		this.data[i][this.w-1] = '|'
		this.data[i][11] = '|'
	}
	this.data[0][0] = '*'
	this.data[0][this.w-1] = '*'
	this.data[this.h-1][0] = '*'
	this.data[this.h-1][this.w-1] = '*'
	this.data[0][11] = '*'
	this.data[2][11] = '*'
	this.data[this.h-1][11] = '*'
	this.data[1][this.w-20] = '|'
	line := []byte("friends")
	for i := 0; i < len(line); i++ {
		this.data[1][1+i] = line[i]
	}
}

func (this *Screen) RenderNames(names []string, selected int) {
	for i := 0; i < len(names); i++ {
		bytes := []byte(names[i])
		offset := 0
		if i == selected {
			offset = 1
			this.data[3+i][0+1] = '^'
		}
		for j := 0; j < len(bytes); j++ {
			this.data[3+i][j+1+offset] = bytes[j]
		}
	}
	line := []byte("chatting with (" + names[selected] + ")")
	for i := 0; i < len(line); i++ {
		this.data[1][40+i] = line[i]
	}
}

func (this *Screen) RenderUnseen(count int) {
	line := []byte("unseen:(" + strconv.Itoa(count) + ")")
	for i := 0; i < len(line); i++ {
		this.data[1][this.w-13+i] = line[i]
	}
}

func (this *Screen) RenderMessage(messages []string) {
	msgs := messages[0:]
	if len(messages) >= 8 {
		msgs = messages[len(messages)-8:]
	}
	for i := 0; i < len(msgs); i++ {
		line := []byte(msgs[i])
		for j := 0; j < len(line); j++ {
			this.data[3+i][12+j] = line[j]
		}
	}
}

func (this *Screen) Render() {
	for i := 0; i < this.h; i++ {
		for j := 0; j < this.w; j++ {
			fmt.Printf("%c", this.data[i][j])
		}
		fmt.Print("\n")
	}
}

const (
	ACTION_SEND_MESSAGE = "ACTION_SEND_MESSAGE"
	ACTION_RECV_MESSAGE = "ACTION_RECV_MESSAGE"
	ACTION_MARK_SEEN    = "ACTION_MARK_SEEN"
	ACTION_SELECT_TAB   = "ACTION_SELECT_TAB"
)

type ThreadStoreData struct {
	Selected int
	Threads  []string
}

type MessageStoreData struct {
	ThreadToMessages map[int][]string
}

type UnSeenStoreData struct {
	UnseenMessage map[int]int // 某个人有多少个未读消息
}

func (this *UnSeenStoreData) Count() int {
	sum := 0
	for _, v := range this.UnseenMessage {
		sum += v
	}
	return sum
}

func main() {
	// https://www.youtube.com/watch?v=nYkdrAPrdcw 依据此视频设计
	// 本文实现其中的Main Message，不实现ChatTab
	// 通过controller发出的Action本质上与waitfor等价

	messageStore := flux.NewStore("Message", MessageStoreData{
		ThreadToMessages: make(map[int][]string),
	}, map[string]flux.Handler{
		ACTION_SEND_MESSAGE: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			args := action.Payload.([]any)
			mapping := store.Data.(MessageStoreData).ThreadToMessages
			mapping[args[0].(int)] = append(mapping[args[0].(int)], args[1].(string))
		},
		ACTION_RECV_MESSAGE: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			args := action.Payload.([]any)
			mapping := store.Data.(MessageStoreData).ThreadToMessages
			mapping[args[0].(int)] = append(mapping[args[0].(int)], args[1].(string))
		},
	}, []flux.Controller{
		func(dispatcher flux.Dispatcher, store *flux.Store, data any) {
			// 这里不允许用同步方法，因为会产生递归调用，其中的visit和pending会出错
			dispatcher.(*flux.Flux).Dispatch(flux.Action{ACTION_MARK_SEEN,
				dispatcher.(*flux.Flux).Stores()["Thread"].Data.(*ThreadStoreData).Selected,
			})
		},
	})
	threadStore := flux.NewStore("Thread", &ThreadStoreData{
		Selected: 0,
		Threads:  []string{"Alice", "Bob", "Peter", "Emma", "Olivia", "Ava", "Isabella"},
	}, map[string]flux.Handler{
		ACTION_SELECT_TAB: func(dispatcher flux.Dispatcher, store *flux.Store, action flux.Action) {
			store.Data.(*ThreadStoreData).Selected = action.Payload.(int)
			dispatcher.(*flux.Flux).Dispatch(flux.Action{ACTION_MARK_SEEN,
				action.Payload.(int),
			})
		},
	}, nil)
	unseenCountStore := flux.NewStore("UnSeen", &UnSeenStoreData{
		UnseenMessage: make(map[int]int),
	}, map[string]flux.Handler{
		ACTION_RECV_MESSAGE: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			args := action.Payload.([]any)
			store.Data.(*UnSeenStoreData).UnseenMessage[args[0].(int)] += 1
		},
		ACTION_MARK_SEEN: func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			store.Data.(*UnSeenStoreData).UnseenMessage[action.Payload.(int)] = 0
		},
	}, nil)
	dispatcher := flux.NewFlux(&flux.AbstractActionCreator{}, messageStore, threadStore, unseenCountStore)

	screen := NewScreen(100, 12)
	screen.Clear()
	screen.RenderNames(threadStore.Data.(*ThreadStoreData).Threads, threadStore.Data.(*ThreadStoreData).Selected)
	screen.RenderUnseen(unseenCountStore.Data.(*UnSeenStoreData).Count())
	screen.RenderMessage(messageStore.Data.(MessageStoreData).ThreadToMessages[threadStore.Data.(*ThreadStoreData).Selected])
	screen.Render()
	tip := ""
	for {
		if tip != "" {
			fmt.Print(tip)
			tip = ""
		}
		fmt.Print("type command (? for help) > ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if err != nil {
			fmt.Printf("读取错误: %v\n", err)
		}
		if input == "?" {
			tip =
				"1. ?                       help\n" +
					"2. send <message>          send message to current tab \n" +
					"3. recv <name> <message>   receive a message by name \n" +
					"4. selt <name>             select user to talk \n" +
					"5. exit                    exit system\n"
		} else if input == "exit" {
			fmt.Println("exited")
			break
		} else {
			frags := strings.Split(input, " ")
			if frags[0] == "send" && len(frags) >= 2 {
				frags2 := strings.SplitN(input, " ", 2)
				dispatcher.Dispatch(
					flux.Action{
						ACTION_SEND_MESSAGE,
						[]any{
							threadStore.Data.(*ThreadStoreData).Selected, "You: " + frags2[1],
						},
					},
				)
				tip = "processed send\n"
			} else if frags[0] == "recv" && len(frags) >= 3 {
				id := -1
				for i := 0; i < len(threadStore.Data.(*ThreadStoreData).Threads); i++ {
					if threadStore.Data.(*ThreadStoreData).Threads[i] == frags[1] {
						id = i
						break
					}
				}
				if id == -1 {
					tip = "name is not exists\n"
				} else {
					frags2 := strings.SplitN(input, " ", 3)
					dispatcher.Dispatch(
						flux.Action{
							ACTION_RECV_MESSAGE,
							[]any{
								id, frags2[1] + ": " + frags2[2],
							},
						},
					)
					tip = "processed recv\n"
				}
			} else if frags[0] == "selt" && len(frags) == 2 {
				id := -1
				for i := 0; i < len(threadStore.Data.(*ThreadStoreData).Threads); i++ {
					if threadStore.Data.(*ThreadStoreData).Threads[i] == frags[1] {
						id = i
						break
					}
				}
				if id == -1 {
					tip = "selcted name is not exists\n"
				} else {
					dispatcher.Dispatch(flux.Action{ACTION_SELECT_TAB, id})
				}
			} else {
				tip = "invalid\n"
			}
		}
		dispatcher.WaitSync()
		time.Sleep(100 * time.Millisecond)
		screen.Clear()
		screen.RenderNames(threadStore.Data.(*ThreadStoreData).Threads, threadStore.Data.(*ThreadStoreData).Selected)
		screen.RenderUnseen(unseenCountStore.Data.(*UnSeenStoreData).Count())
		screen.RenderMessage(messageStore.Data.(MessageStoreData).ThreadToMessages[threadStore.Data.(*ThreadStoreData).Selected])
		screen.Render()
	}
}
