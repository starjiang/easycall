package easycall

//Response for EasyService
type Response struct {
	format byte
	head   *EasyHead
	body   interface{}
}

func (r *Response) GetBody() interface{} {
	return r.body
}

func (r *Response) GetHead() *EasyHead {
	return r.head
}
func (r *Response) GetFormat() byte {
	return r.format
}

func (r *Response) SetBody(body interface{}) *Response {

	r.body = body
	return r
}

func (r *Response) SetHead(head *EasyHead) *Response {

	r.head = head
	return r
}

func (r *Response) SetFormat(format byte) *Response {

	r.format = format
	return r
}
