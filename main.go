package main

import (
	"testing"

	"gorm-unit-test-demo/repo/test"
)

func main() {
	t := testing.T{}
	test.SetupLocalDatabaseConnection(&t, "")
}
