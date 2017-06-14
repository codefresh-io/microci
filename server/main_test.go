package main

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

//---- TESTS

func Test_main(t *testing.T) {
	os.Args = []string{"microci", "-v"}
	main()
}

func Test_handleSignals_SIGTERM(t *testing.T) {
	sigs := make(chan os.Signal, 1)
	handleSignals(sigs, false)
	// send SIGTERM signal
	sigs <- syscall.SIGTERM
	// wait a while to handle signal
	time.Sleep(time.Millisecond)
}

func Test_dummy(t *testing.T) {
	fmt.Println("Hello")
}
