package sched

import (
	"testing"

	"zakirullin/stuffbot/pkg/txt"

	"github.com/stretchr/testify/require"
)

func TestUcfirst(t *testing.T) {
	r := require.New(t)

	res := txt.Ucfirst("abc")

	r.Equal("Abc", res)
}

func TestUcfirstRu(t *testing.T) {
	r := require.New(t)

	res := txt.Ucfirst("абв")

	r.Equal("Абв", res)
}
