package agent

// Linux installations use a user-supplied supervisor; do not guess its unit name.
func prepareServiceUpdate(exe, config string) (serviceUpdate, error) {
	return serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}, nil
}
