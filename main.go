package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

func probeMessagePassing() {
	R1 := make(chan int, 1)
	R2 := make(chan int, 1)

	var x, y int

	f1 := func() {
		x = 1
		y = 1
	}

	f2 := func() {
		r1 := y
		r2 := x
		R1 <- r1
		R2 <- r2
	}

	run(f1, f2)

	if <-R1 == 1 && <-R2 == 0 {
		log.Printf("message passing detected!")
	}
}

func probeBufferedWrites() {
	R1 := make(chan int, 1)
	R2 := make(chan int, 1)

	var x, y int

	f1 := func() {
		x = 1
		r1 := y
		R1 <- r1
	}

	f2 := func() {
		y = 1
		r2 := x
		R2 <- r2
	}

	run(f1, f2)

	if <-R1 == 0 && <-R2 == 0 {
		log.Printf("write buffering detected!")
	}
}

func probeIRIW() {
	R1 := make(chan int, 1)
	R2 := make(chan int, 1)
	R3 := make(chan int, 1)
	R4 := make(chan int, 1)

	var x, y int

	f1 := func() {
		x = 1
	}

	f2 := func() {
		y = 1
	}

	f3 := func() {
		r1 := x
		r2 := y
		R1 <- r1
		R2 <- r2
	}

	f4 := func() {
		r3 := y
		r4 := x
		R3 <- r3
		R4 <- r4
	}

	run(f1, f2, f3, f4)

	if <-R1 == 1 && <-R2 == 0 && <-R3 == 1 && <-R4 == 0 {
		log.Printf("iriw: detected!")
	}
}

func probeN6() {
	R1 := make(chan int, 1)
	R2 := make(chan int, 1)

	var x, y int

	f1 := func() {
		x = 1
		r1 := x
		r2 := y
		R1 <- r1
		R2 <- r2
	}

	f2 := func() {
		y = 1
		x = 2
	}

	run(f1, f2)

	if <-R1 == 1 && <-R2 == 0 && x == 1 {
		log.Printf("n6: detected!")
	}
}

func probeReadBuffering() {
	R1 := make(chan int, 1)
	R2 := make(chan int, 1)

	var x, y int

	f1 := func() {
		r1 := x
		y = 1
		R1 <- r1
	}

	f2 := func() {
		r2 := y
		x = 1
		R2 <- r2
	}

	run(f1, f2)

	if <-R1 == 1 && <-R2 == 1 {
		log.Printf("read buffering detected!")
	}
}

// run ensures that the provided functions are run in a synchronized way.
func run(fs ...func()) {
	// shuffle the order we run the goroutines
	perm := rand.Perm(len(fs))

	// setup wait groups to ensure all go routines are started before
	// running funcs and all funcs finish before checking results.
	var wgStart, wgDone sync.WaitGroup
	wgStart.Add(len(fs))
	wgDone.Add(len(fs))

	// create and wait for all goroutines to start
	for _, i := range perm {
		go func(f func()) {
			wgStart.Done()
			wgStart.Wait()
			f()
			wgDone.Done()
		}(fs[i])
	}

	wgDone.Wait()
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "probe-memory-model",
		Short: "A tool for running various memory model probes in Go.",
		Long: `This is a tool which implements a few of the memory model probes
discussed in Russ Cox's "Hardware Memory Models" blog post.

Note that variables are initialized to zero in all probes.
	`,
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "mp",
		Short: "Probe for message passing.",
		Long: `This probe runs the following test to determine whether message passing is happening:

Proc 1        Proc 2
x = 1         r1 = y
y = 1         r2 = x

Can we see?
r1 = 1
r2 = 0
`,
		Run: func(cmd *cobra.Command, args []string) {
			for {
				probeMessagePassing()
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "bw",
		Short: "Probe for buffered writes.",
		Long: `This probe runs the following test to determine whether writes are buffered in a queue:
		
Proc 1        Proc 2
 x = 1         y = 1
r1 = y        r2 = x

Can we see?
r1 = 0
r2 = 0
`,
		Run: func(cmd *cobra.Command, args []string) {
			for {
				probeBufferedWrites()
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "iriw",
		Short: "Probe for independent reads of independent writes.",
		Long: `This probe runs the following test to determine whether independent reads occur for independent writes:
		
Proc 1        Proc 2        Proc 3        Proc 3
x = 1         y = 1         r1 = x        r3 = y
                            r2 = y        r4 = x

Can we see?
r1 = 1
r2 = 0
r3 = 1
r4 = 0
`,
		Run: func(cmd *cobra.Command, args []string) {
			for {
				probeIRIW()
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "n6",
		Short: "Probe by Paul Loewenstein to show x86 violates TLO+CC memory model.",
		Long: `This probe runs the following test to determine whether memory write queues:
		
Proc 1        Proc 2
 x = 1        y = 1
r1 = x        x = 2
r2 = y

Can we see?
r1 = 1
r2 = 0
 x = 1
`,
		Run: func(cmd *cobra.Command, args []string) {
			for {
				probeN6()
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "rb",
		Short: "Probe for read buffering.",
		Long: `This probe runs the following test to determine whether read buffering can happen:
		
Proc 1        Proc 2
r1 = x        r2 = y
 y = 1         x = 1

Can we see?
r1 = 1
r2 = 1
`,
		Run: func(cmd *cobra.Command, args []string) {
			for {
				probeReadBuffering()
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
