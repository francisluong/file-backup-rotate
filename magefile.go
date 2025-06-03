// +build mage

package main

import (
    "fmt"
    "os"
    "os/exec"
)

// Test runs all Go tests in the repository.
func Test() error {
    cmd := exec.Command("go", "test", "./...")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    fmt.Println("Running all tests...")
    return cmd.Run()
}
