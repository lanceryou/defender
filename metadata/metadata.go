package metadata

import (
	"context"
	"fmt"
	"strings"
)

type mdKey struct{}

// MD metadata
type MD map[string]string

// New metadata
func New(m map[string]string) MD {
	md := MD{}
	for k, val := range m {
		key := strings.ToLower(k)
		md[key] = val
	}
	return md
}

// Pairs
func Pairs(kv ...string) MD {
	if len(kv)%2 == 1 {
		panic(fmt.Errorf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kv)))
	}
	md := MD{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			key = strings.ToLower(s)
			continue
		}
		md[key] = s
	}
	return md
}

// Join joins any number of mds into a single MD.
// The order of values for each key is determined by the order in which
// the mds containing those values are presented to Join.
func Join(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out[k] = v
		}
	}
	return out
}

// Len returns the number of items in md.
func (md MD) Len() int {
	return len(md)
}

// Copy returns a copy of md.
func (md MD) Copy() MD {
	return Join(md)
}

// Set sets the value of a given key with value.
func (m MD) Set(key string, val string) {
	m[key] = val
}

// Get obtains the values for a given key.
func (m MD) Get(key string) string {
	return m[key]
}

// NewMetadataFromContext creates a new context with md attached.
func NewMetadataFromContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdKey{}, md)
}

// FromContext
func FromContext(ctx context.Context) (md MD) {
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return nil
	}
	return
}
