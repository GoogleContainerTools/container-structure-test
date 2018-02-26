package drivers

type ContainerRunOpts struct {
	BindMounts   []string `yaml:"bindMounts"`
	Privileged   bool     `yaml:"privileged"`
	Capabilities []string `yaml:"capabilities"`
}
