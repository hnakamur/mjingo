package mjingo

type TestFunc = func(*virtualMachineState, []value) (bool, error)

func isUndefined(val value) bool {
	return val.isUndefined()
}
