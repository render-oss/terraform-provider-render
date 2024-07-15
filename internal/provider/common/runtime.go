package common

import "terraform-provider-render/internal/client"

func ToClientRuntime(runtime string) client.ServiceRuntime {
	switch runtime {
	case "docker":
		return client.ServiceRuntimeDocker
	case "image":
		return client.ServiceRuntimeImage
	case "node":
		return client.ServiceRuntimeNode
	case "python":
		return client.ServiceRuntimePython
	case "ruby":
		return client.ServiceRuntimeRuby
	case "rust":
		return client.ServiceRuntimeRust
	case "go":
		return client.ServiceRuntimeGo
	case "elixir":
		return client.ServiceRuntimeElixir
	}

	return ""
}
