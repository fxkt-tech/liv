package naming

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Naming struct {
	// names map[string]bool
}

func New() *Naming {
	return &Naming{}
}

func (n *Naming) Gen() string {
	return fmt.Sprintf("%x", rand.Int31n(math.MaxInt32))
}

func (n *Naming) Gen64() string {
	return fmt.Sprintf("%x", math.MaxInt64)
}
