package common

// InstanceConfig ...
type InstanceConfig struct {
	ID             string
	Name           string
	Region         string
	Zone           string
	Image          string
	Size           string
	InternalIP     string
	ExteralIP      string
	SSHFingerprint string
	Tags           []string
	Labels         map[string]string
	IsRunning      bool
}
