package easycall

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/vmihailenco/msgpack"
)

//Request for EasyService
type Request struct {
	format        byte
	head          *EasyHead
	bodyData      []byte
	createTime    time.Time
	remoteAddress string
}

func (r *Request) GetBody(body interface{}) error {

	if r.format == FORMAT_MSGPACK {
		return msgpack.NewDecoder(bytes.NewReader(r.bodyData)).UseJSONTag(true).Decode(body)
	} else if r.format == FORMAT_JSON {
		return json.NewDecoder(bytes.NewReader(r.bodyData)).Decode(body)
	} else {
		return errors.New("invalid package format")
	}
}

func (r *Request) GetBodyData() []byte {
	return r.bodyData
}

func (r *Request) GetHead() *EasyHead {
	return r.head
}

func (r *Request) GetCreateTime() time.Time {
	return r.createTime
}

func (r *Request) GetRemoteAddress() string {
	return r.remoteAddress
}

func (r *Request) GetFormat() byte {
	return r.format
}
