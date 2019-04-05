package llif

type ExecCreateInput struct {
	NetworkOutput NetworkCreateOutput
	FSOutput      FSCreateOutput
}

type ExecCreateOutput struct {
	// FsOpt will be passed to the ExecHandler
	FsOpt map[string]string
}

type ExecRunInput struct {
	NetworkOutput NetworkRunOutput
	FSOutput      FSRunOutput
}

type ExecRunOutput struct {
	FsOpt map[string]string
}

type ExecDestroyInput struct {
}

// ExecDestroyOutput will be passed on to FSDestroyFunc and NetworkDestroyFunc
type ExecDestroyOutput struct {
	FsOpt map[string]string
}
