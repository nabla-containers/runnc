package llif

type NetworkGenericInput struct {
	// The state of LL handlers
	FSState      *LLState
	NetworkState *LLState
	ExecState    *LLState
}

type NetworkCreateInput struct {
	NetworkGenericInput
}

type NetworkRunInput struct {
	NetworkGenericInput
}

type NetworkDestroyInput struct {
	NetworkGenericInput
}
