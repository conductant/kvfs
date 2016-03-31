package kvfs

type ErrNotSupported struct {
	Protocol string
}

func (this *ErrNotSupported) Error() string {
	return "Protocol not supported:" + this.Protocol
}
