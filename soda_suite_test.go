package soda_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSoda(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Soda Suite")
}
