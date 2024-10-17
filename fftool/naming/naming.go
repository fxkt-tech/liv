package naming

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var Default = New()

type Naming struct {
	rand *rand.Rand
}

func New() *Naming {
	return &Naming{
		rand: rand.New(rand.NewSource((time.Now().UnixNano()))),
	}
}

func (n *Naming) Gen() string {
	return fmt.Sprintf("%x", n.rand.Int31n(math.MaxInt32))
}

func (n *Naming) Gen64() string {
	return fmt.Sprintf("%x", n.rand.Int63n(math.MaxInt64))
}

func (n *Naming) Empty() string {
	return ""
}
