package llif

type ExecGenericInput struct {
	// The state of LL handlers
	FSState      *LLState
	NetworkState *LLState
	ExecState    *LLState
}

type ExecCreateInput struct {
	ExecGenericInput
}

type ExecRunInput struct {
	ExecGenericInput
}

type ExecDestroyInput struct {
	ExecGenericInput
}
