package volume

import (
	"github.com/docker/docker/api/types/mount"
	"strings"
)

func (v *Volume) GetMountType() mount.Type {
	if v.Type != "" {
		return mount.Type(v.Type)
	}

	if strings.HasPrefix(v.Source, "/") {
		return mount.TypeBind
	}

	return mount.TypeVolume
}

func (v *Volume) IsReadOnly() bool {
	if v.ReadOnly {
		return true
	}

	for _, mode := range strings.Split(v.Mode, ",") {
		if mode == "ro" {
			return true
		}
	}

	return false
}
