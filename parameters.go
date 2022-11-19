package gobatch

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/karlseguin/typed"
)

type Parameters struct {
	typed.Typed
}

func (p *Parameters) Set(k string, v any) *Parameters {
	p.Typed[k] = v
	return p
}

func (p Parameters) ToString() string {
	bs, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

func (p *Parameters) FromString(str string) error {
	return json.Unmarshal([]byte(str), p)
}

func (p *Parameters) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &p.Typed)
}

func (p Parameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Typed)
}

func (p Parameters) Footprint() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	b := md5.Sum(bytes)
	return fmt.Sprintf("%x", b)
}
