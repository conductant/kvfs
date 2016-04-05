package kvfs

type ErrFailedDelete struct {
	Key string
}

func (this *ErrFailedDelete) Error() string {
	return "Failed to delete key=" + this.Key
}

type ErrNotSupported struct {
	Protocol string
}

func (this *ErrNotSupported) Error() string {
	return "Protocol not supported:" + this.Protocol
}
