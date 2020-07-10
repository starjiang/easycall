package easycall

type PkgHandler interface {
	Dispatch(pkgData []byte, client *EasyConnection)
}
