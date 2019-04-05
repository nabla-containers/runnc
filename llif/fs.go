package llif

type FSCreateInput struct {
}

// FSCreateOutput will be passed on to ExecCreateFunc
type FSCreateOutput struct {
	FsOpt map[string]string
}

type FSRunInput struct {
}

// FSRunOutput will be passed on to ExecRunFunc
type FSRunOutput struct {
	FsOpt map[string]string
}

type FSDestroyInput struct {
	ExecOutput ExecDestroyOutput
}

type FSDestroyOutput struct {
	FsOpt map[string]string
}
