package easycall

import "github.com/starjiang/elog"

//ClientHandler for EasyService
type ClientHandler struct {
	client interface{}
}

func NewClientHandler(client interface{}) *ClientHandler {
	ClientHandler := &ClientHandler{}
	ClientHandler.client = client
	return ClientHandler
}

func (h *ClientHandler) Dispatch(pkgData []byte, client *EasyConnection) {
	serviceClient := h.client.(*ServiceClient)
	reqPkg, err := DecodeWithBodyData(pkgData)
	if err != nil {
		elog.Error("decode pkg fail:", err)
	}
	serviceClient.Process(reqPkg)
}
