package engine

import (
	"strings"

	"github.com/docker/docker/api/types/mount"
)

func (v *Volume) getMountType() mount.Type {
	if v.Type != "" {
		return mount.Type(v.Type)
	}

	if strings.HasPrefix(v.Source, "/") {
		return mount.TypeBind
	}

	return mount.TypeVolume
}

func (v *Volume) isReadOnly() bool {
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
