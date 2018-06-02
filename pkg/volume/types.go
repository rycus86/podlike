package volume

type Volume struct {
	Type     string
	Source   string
	Target   string
	ReadOnly bool `mapstructure:"read_only"`
	Mode     string

	Bind struct {
		Propagation string
	}

	Volume struct {
		NoCopy bool `mapstructure:"nocopy"`
	}

	Tmpfs struct {
		Size string
	}
}
