package main

// +build linux

import(
	// only used for output
	"fmt"
	// only used for os.Exit
	"os"

	seccomp "github.com/seccomp/libseccomp-golang"
)

// simple cli based demo for filtering syscalls. uses a whitelist.
// demonstrates how to filter and what happens when a not allowed syscall is executed.
func main() {
	// for allowing, the actual name of the syscall is needed
	syscallAllowedActualName := "getpid"
	// store the actual result before in order to compare later
	// this is just for showing how it works
	actualAllowedResult := os.Getpid()
	fmt.Printf("current result of syscall '%v': %v\n", syscallAllowedActualName,
			actualAllowedResult)

	// for testing, before applying the filter, store the actual result
	// of a function later not to be allowed
	syscallNotAllowedActualName := "getgid"
	actualNotAllowedResult := os.Getgid()
	fmt.Printf("current result of syscall '%v': %v\n", syscallNotAllowedActualName,
			actualNotAllowedResult)

	// create a new filter with a default Action
	// ActErrno will make the syscall return with error
	// ActKill would kill the program with SIGSYS
	filter, err := seccomp.NewFilter(seccomp.ActErrno)
	// filter, err := seccomp.NewFilter(seccomp.ActKill)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating seccomp filter: %v\n", err)
		os.Exit(1)
	}


	// allow required default syscalls by the go runtime
	// for a real world program, you will probably need to allow more syscalls
	// use `strace -CDf` to find out which syscalls you will probably need

	// used by all runtime components
	// man 2 futex - fast user-space locking
	allowSyscall(filter, "futex")

	// used by the garbage collector
	// man 2 mmap, munmap - map or unmap files or devices into memory
	allowSyscall(filter, "munmap")

	// used to exit the program when done
	// man 2 exit_group - exit all threads in a process
	allowSyscall(filter, "exit_group")


	// allow required write function, otherwise fmt.Print* will not work.
	// if you do not use fmt.Print* with an external resource, it might not be required.
	// man 2 write - write to a file descriptor
	allowSyscall(filter, "write")


	// now allow the above defined syscall
	allowSyscall(filter, syscallAllowedActualName)


	// apply the filter
	// when this is done without error, no more adjustments can be made to seccomp
	// and no other syscalls other then the allowed ones succeed!
	err = filter.Load(); if err != nil {
		fmt.Fprintf(os.Stderr, "error loading seccomp: %v\n", err)
		os.Exit(3)
	}

	fmt.Printf("created seccomp whitelist filter with default action Errno. allowed functions: '%v'\n",
			syscallAllowedActualName)

	// execute syscalls again, with filter enabled
	resultAllowed := os.Getpid()
	resultNotAllowed := os.Getgid()

	// if this happens, the filter failed and you encountered a bug. no real world code, just for testing.
	if resultNotAllowed == actualNotAllowedResult {
		fmt.Printf("unexpected, probably bug: result of syscall '%v' is %v instead of 0 or negative\n" +
				"result of syscall '%v': %v",
				syscallNotAllowedActualName, resultNotAllowed,
				syscallAllowedActualName, resultAllowed)
		os.Exit(4)
	}

	// present what happend
	fmt.Printf("current result of syscall '%v': %v\ncurrent result of syscall '%v': %v\n",
				syscallAllowedActualName, resultAllowed,
				syscallNotAllowedActualName, resultNotAllowed)
}

// allow a syscall. retrieves the syscall name via seccomp.GetSyscallFromName
// which returns the syscall number for the current architecture and then adds
// the ActAllow rule for it.
func allowSyscall(filter *seccomp.ScmpFilter, name string) {
	allowed, err := seccomp.GetSyscallFromName(name); if err != nil {
		fmt.Fprintf(os.Stderr, "error getting syscall number on %v: %v\n", name, err)
		os.Exit(9)
	}

	filter.AddRule(allowed, seccomp.ActAllow)
}