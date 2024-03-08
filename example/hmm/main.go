package main

import (
	"fmt"
	"github.com/clouddea/flux-go/flux"
	"math/rand"
)

type StoreData struct {
	s int // 隐状态
	o int // 显状态
}

func randomSelectElementByPossibility(p []float64) int {
	randomPos := rand.Float64()
	sum := 0.0
	for i := 0; i < len(p); i++ {
		sum += p[i]
		if randomPos < sum {
			return i
		}
	}
	return len(p) - 1
}

func print2DArrayWithPrecision(array [][]float64, precision int) {
	for _, row := range array {
		for _, value := range row {
			fmt.Printf("%.*f ", precision, value)
		}
		fmt.Println()
	}
}

func main() {
	nHiddenStates := 8
	nObservedStates := 10
	A := make([][]float64, nHiddenStates)
	AA := make([][]float64, nHiddenStates)
	for i := 0; i < nHiddenStates; i++ {
		A[i] = make([]float64, nHiddenStates)
		AA[i] = make([]float64, nHiddenStates)
		sum := 0.0
		for j := 0; j < nHiddenStates; j++ {
			A[i][j] = rand.Float64()
			sum += A[i][j]
		}
		for j := 0; j < nHiddenStates; j++ {
			A[i][j] /= sum
		}
	}

	B := make([][]float64, nHiddenStates)
	BB := make([][]float64, nHiddenStates)
	for i := 0; i < nHiddenStates; i++ {
		B[i] = make([]float64, nObservedStates)
		BB[i] = make([]float64, nObservedStates)
		sum := 0.0
		for j := 0; j < nObservedStates; j++ {
			B[i][j] = rand.Float64()
			sum += B[i][j]
		}
		for j := 0; j < nObservedStates; j++ {
			B[i][j] /= sum
		}
	}

	pi := make([]float64, nHiddenStates)
	sum := 0.0
	for i := 0; i < nHiddenStates; i++ {
		pi[i] = rand.Float64()
		sum += pi[i]
	}
	for i := 0; i < nHiddenStates; i++ {
		pi[i] /= sum
	}

	fmt.Printf("A=\n")
	print2DArrayWithPrecision(A, 3)
	fmt.Printf("B=\n")
	print2DArrayWithPrecision(B, 3)
	//fmt.Printf("pi=%v\n", pi)
	//print2DArrayWithPrecision(pi, 3)
	// start simulation
	hiddenStateStore := flux.NewStore("hiddenState", 0, map[string]flux.Handler{
		"INIT": func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			store.Data = randomSelectElementByPossibility(pi)
			//fmt.Println(1)
		},
		"TRANSFER": func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			store.Data = randomSelectElementByPossibility(A[store.Data.(int)])
			//fmt.Println("a")
		},
	}, nil)

	observedStateStore := flux.NewStore("observedState", 0, map[string]flux.Handler{
		"INIT": func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			flux.WaitFor("hiddenState")
			store.Data = randomSelectElementByPossibility(B[hiddenStateStore.Data.(int)])
			//fmt.Println(2)
		},
		"TRANSFER": func(flux flux.Dispatcher, store *flux.Store, action flux.Action) {
			flux.WaitFor("hiddenState")
			store.Data = randomSelectElementByPossibility(B[hiddenStateStore.Data.(int)])
			//fmt.Println("b")
		},
	}, nil)

	dispatcher := flux.NewFlux(&flux.AbstractActionCreator{}, observedStateStore, hiddenStateStore)
	dispatcher.DispatchSync(flux.Action{"INIT", nil})
	last := hiddenStateStore.Data.(int)

	for i := 0; i < 1000000; i++ {
		dispatcher.DispatchSync(flux.Action{"TRANSFER", nil})
		AA[last][hiddenStateStore.Data.(int)] += 1
		last = hiddenStateStore.Data.(int)
		BB[hiddenStateStore.Data.(int)][observedStateStore.Data.(int)] += 1
	}
	// recal A and B
	for i := 0; i < nHiddenStates; i++ {
		sum := 0.0
		for j := 0; j < nHiddenStates; j++ {
			sum += AA[i][j]
		}
		for j := 0; j < nHiddenStates; j++ {
			AA[i][j] /= sum
		}
	}
	for i := 0; i < nHiddenStates; i++ {
		sum := 0.0
		for j := 0; j < nObservedStates; j++ {
			sum += BB[i][j]
		}
		for j := 0; j < nObservedStates; j++ {
			BB[i][j] /= sum
		}
	}

	fmt.Printf("AA=\n")
	print2DArrayWithPrecision(AA, 3)
	fmt.Printf("BB=\n")
	print2DArrayWithPrecision(BB, 3)

}
