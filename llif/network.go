package llif

type NetworkCreateInput struct {
}

// NetworkCreateOutput will be passed on to ExecCreateFunc
type NetworkCreateOutput struct {
	FsOpt map[string]string
}

type NetworkRunInput struct {
}

// NetworkRunOutput will be passed on to ExecRunFunc
type NetworkRunOutput struct {
	FsOpt map[string]string
}

type NetworkDestroyInput struct {
	ExecOutput ExecDestroyOutput
}

type NetworkDestroyOutput struct {
	FsOpt map[string]string
}
