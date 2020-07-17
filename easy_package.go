package easycall

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/vmihailenco/msgpack"
)

const (
	STX            = 0x2
	ETX            = 0x3
	HEAD_MAX_LEN   = 128 * 2014
	BODY_MAX_LEN   = 2 * 1024 * 1024
	FORMAT_JSON    = 1
	FORMAT_MSGPACK = 0
)

//EasyHead for EasyPackage
type EasyHead struct {
	Service   string `json:"service"`
	Method    string `json:"method"`
	RouteKey  string `json:"routeKey"`
	Token     string `json:"token"`
	Uid       uint64 `json:"uid"`
	RequestIp string `json:"requestIp"`
	TraceId   string `json:"traceId"`
	Seq       uint64 `json:"seq"`
	Ret       int    `json:"ret"`
	Msg       string `json:"msg"`
}

func NewEasyHead() *EasyHead {
	return &EasyHead{}
}

func (head *EasyHead) GetService() string {
	return head.Service
}

func (head *EasyHead) GetMethod() string {
	return head.Method
}
func (head *EasyHead) GetRouteKey() string {
	return head.RouteKey
}
func (head *EasyHead) GetToken() string {
	return head.Token
}
func (head *EasyHead) GetUid() uint64 {
	return head.Uid
}
func (head *EasyHead) GetRequestIp() string {
	return head.RequestIp
}
func (head *EasyHead) GetTraceId() string {
	return head.TraceId
}
func (head *EasyHead) GetSeq() uint64 {
	return head.Seq
}

func (head *EasyHead) GetRet() int {
	return head.Ret
}

func (head *EasyHead) GetMsg() string {
	return head.Msg
}

func (head *EasyHead) SetService(service string) *EasyHead {
	head.Service = service
	return head
}

func (head *EasyHead) SetMethod(method string) *EasyHead {
	head.Method = method
	return head
}

func (head *EasyHead) SetRouteKey(routeKey string) *EasyHead {
	head.RouteKey = routeKey
	return head
}

func (head *EasyHead) SetToken(token string) *EasyHead {
	head.Token = token
	return head
}

func (head *EasyHead) SetUid(uid uint64) *EasyHead {
	head.Uid = uid
	return head
}

func (head *EasyHead) SetRequestIp(requestIp string) *EasyHead {
	head.RequestIp = requestIp
	return head
}

func (head *EasyHead) SetTraceId(traceId string) *EasyHead {
	head.TraceId = traceId
	return head
}

func (head *EasyHead) SetSeq(seq uint64) *EasyHead {
	head.Seq = seq
	return head
}

func (head *EasyHead) SetRet(ret int) *EasyHead {
	head.Ret = ret
	return head
}

func (head *EasyHead) SetMsg(msg string) *EasyHead {
	head.Msg = msg
	return head
}

//EasyPackage for Easycall
type EasyPackage struct {
	format   byte
	head     *EasyHead
	bodyData []byte
	pkgData  []byte
	body     interface{}
}

func NewPackageWithBodyData(format byte, head *EasyHead, bodyData []byte) *EasyPackage {
	return &EasyPackage{format, head, bodyData, nil, nil}
}

func NewPackageWithBody(format byte, head *EasyHead, body interface{}) *EasyPackage {
	return &EasyPackage{format, head, nil, nil, body}
}

func DecodeWithBodyData(pkgData []byte) (*EasyPackage, error) {

	var err error = nil

	format := pkgData[1]
	headLen := binary.BigEndian.Uint32(pkgData[2:6])
	bodyLen := binary.BigEndian.Uint32(pkgData[6:10])

	headData := pkgData[10 : 10+headLen]
	bodyData := pkgData[10+headLen : 10+headLen+bodyLen]
	if format == FORMAT_JSON {
		head := &EasyHead{}
		err = json.NewDecoder(bytes.NewReader(headData)).Decode(head)
		if err != nil {
			return nil, err
		}
		return &EasyPackage{format, head, bodyData, pkgData, nil}, nil

	} else if format == FORMAT_MSGPACK {
		head := &EasyHead{}
		err = msgpack.NewDecoder(bytes.NewReader(headData)).UseJSONTag(true).Decode(head)
		if err != nil {
			return nil, err
		}
		return &EasyPackage{format, head, bodyData, pkgData, nil}, nil
	} else {
		return nil, errors.New("invalid pkg format")
	}
}

func DecodeWithBody(pkgData []byte) (*EasyPackage, error) {

	var err error = nil

	format := pkgData[1]
	headLen := binary.BigEndian.Uint32(pkgData[2:6])
	bodyLen := binary.BigEndian.Uint32(pkgData[6:10])

	headData := pkgData[10 : 10+headLen]
	bodyData := pkgData[10+headLen : 10+headLen+bodyLen]
	if format == FORMAT_JSON {
		head := &EasyHead{}
		err = json.NewDecoder(bytes.NewReader(headData)).Decode(head)
		if err != nil {
			return nil, err
		}
		body := make(map[string]interface{}, 0)
		err = json.NewDecoder(bytes.NewReader(bodyData)).Decode(&body)
		if err != nil {
			return nil, err
		}
		return &EasyPackage{format, head, nil, pkgData, (interface{})(body)}, nil

		return NewPackageWithBody(format, head, (interface{})(body)), nil
	} else if format == FORMAT_MSGPACK {
		head := &EasyHead{}
		err = msgpack.NewDecoder(bytes.NewReader(headData)).UseJSONTag(true).Decode(head)
		if err != nil {
			return nil, err
		}

		body := make(map[string]interface{}, 0)
		err = msgpack.NewDecoder(bytes.NewReader(bodyData)).UseJSONTag(true).Decode(body)
		if err != nil {
			return nil, err
		}
		return &EasyPackage{format, head, nil, pkgData, (interface{})(body)}, nil
	} else {
		return nil, errors.New("invalid pkg format")
	}
}

func (pkg *EasyPackage) EncodeWithBody() ([]byte, error) {

	var headLen int
	var bodyLen int
	var err error

	var buf bytes.Buffer
	prefix := make([]byte, 10)
	buf.Write(prefix)

	if pkg.format == FORMAT_MSGPACK {

		err = msgpack.NewEncoder(&buf).UseJSONTag(true).Encode(pkg.head)
		if err != nil {
			return nil, err
		}
		headLen = buf.Len() - 10
		err = msgpack.NewEncoder(&buf).UseJSONTag(true).Encode(pkg.body)
		if err != nil {
			return nil, err
		}
		bodyLen = buf.Len() - headLen - 10

	} else if pkg.format == FORMAT_JSON {

		err = json.NewEncoder(&buf).Encode(pkg.head)
		if err != nil {
			return nil, err
		}
		headLen = buf.Len() - 10
		err = json.NewEncoder(&buf).Encode(pkg.body)
		if err != nil {
			return nil, err
		}
		bodyLen = buf.Len() - headLen - 10
	} else {
		return nil, errors.New("invalid pkg format")
	}

	buf.WriteByte(ETX)
	pkgData := buf.Bytes()
	pkgData[0] = STX
	pkgData[1] = pkg.format
	binary.BigEndian.PutUint32(pkgData[2:6], uint32(headLen))
	binary.BigEndian.PutUint32(pkgData[6:10], uint32(bodyLen))
	return pkgData, nil

}

func (pkg *EasyPackage) EncodeWithBodyData() ([]byte, error) {

	var headLen int
	var bodyLen int
	var err error

	var buf bytes.Buffer
	prefix := make([]byte, 10)
	buf.Write(prefix)

	if pkg.format == FORMAT_MSGPACK {

		err = msgpack.NewEncoder(&buf).UseJSONTag(true).Encode(pkg.head)
		if err != nil {
			return nil, err
		}
		headLen = buf.Len() - 10
		buf.Write(pkg.bodyData)
		bodyLen = buf.Len() - headLen - 10
	} else if pkg.format == FORMAT_JSON {

		err = json.NewEncoder(&buf).Encode(pkg.head)
		if err != nil {
			return nil, err
		}
		headLen = buf.Len() - 10
		buf.Write(pkg.bodyData)
		bodyLen = buf.Len() - headLen - 10

	} else {
		return nil, errors.New("invalid pkg format")
	}

	buf.WriteByte(ETX)
	pkgData := buf.Bytes()
	pkgData[0] = STX
	pkgData[1] = pkg.format
	binary.BigEndian.PutUint32(pkgData[2:6], uint32(headLen))
	binary.BigEndian.PutUint32(pkgData[6:10], uint32(bodyLen))
	return pkgData, nil

}

func (pkg *EasyPackage) GetFormat() byte {
	return pkg.format
}

func (pkg *EasyPackage) SetFormat(format byte) *EasyPackage {
	pkg.format = format
	return pkg
}

func (pkg *EasyPackage) SetHead(head *EasyHead) *EasyPackage {
	pkg.head = head
	return pkg
}

func (pkg *EasyPackage) SetBody(body interface{}) *EasyPackage {
	pkg.body = body
	return pkg
}

func (pkg *EasyPackage) GetHead() *EasyHead {

	return pkg.head
}

func (pkg *EasyPackage) GetBodyData() []byte {
	return pkg.bodyData
}

func (pkg *EasyPackage) SetBodyData(bodyData []byte) *EasyPackage {
	pkg.bodyData = bodyData
	return pkg
}

func (pkg *EasyPackage) GetPkgData() []byte {
	return pkg.pkgData
}

func (pkg *EasyPackage) SetPkgData(pkgData []byte) *EasyPackage {
	pkg.pkgData = pkgData
	return pkg
}

func (pkg *EasyPackage) GetBody() interface{} {
	return pkg.body
}

func (pkg *EasyPackage) DecodeBody(body interface{}) error {

	if pkg.format == FORMAT_MSGPACK {
		return msgpack.NewDecoder(bytes.NewReader(pkg.bodyData)).UseJSONTag(true).Decode(body)
	} else if pkg.format == FORMAT_JSON {
		return json.NewDecoder(bytes.NewReader(pkg.bodyData)).Decode(body)
	} else {
		return errors.New("invalid package format")
	}
}
