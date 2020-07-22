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
	format        byte                   // request package format 0 for MSGPACK,1 for Json
	head          *EasyHead              //request head struct
	bodyData      []byte                 //request body byte array
	createTime    time.Time              //request create time
	remoteAddress string                 //request remote ip
	ext           map[string]interface{} //for data transmission among middlewares
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

func (r *Request) GetExt() map[string]interface{} {
	return r.ext
}

func (r *Request) GetExtValue(key string) interface{} {
	return r.ext[key]
}

func (r *Request) SetExtValue(key string, value interface{}) {
	r.ext[key] = value
}
