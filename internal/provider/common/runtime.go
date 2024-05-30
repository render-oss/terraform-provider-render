package common

import "terraform-provider-render/internal/client"

func ToClientRuntime(runtime string) client.ServiceEnv {
	switch runtime {
	case "docker":
		return client.ServiceEnvDocker
	case "image":
		return client.ServiceEnvImage
	case "node":
		return client.ServiceEnvNode
	case "python":
		return client.ServiceEnvPython
	case "ruby":
		return client.ServiceEnvRuby
	case "rust":
		return client.ServiceEnvRust
	case "go":
		return client.ServiceEnvGo
	case "elixir":
		return client.ServiceEnvElixir
	}

	return client.ServiceEnv("")
}
