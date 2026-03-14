package modules

func ModuleNames() []string {
	registry := NewRegistry()
	descriptors := registry.Descriptors()
	names := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		names = append(names, descriptor.Name)
	}

	return names
}
