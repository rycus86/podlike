package engine

import (
	"strings"

	"github.com/docker/docker/api/types/mount"
)

func (v *Volume) getMountType() mount.Type {
	if v.Type != "" {
		return mount.Type(v.Type)
	}

	if strings.Index(v.Source, "/") == 0 {
		return mount.TypeBind
	}

	return mount.TypeVolume
}

func (v *Volume) isReadOnly() bool {
	return v.ReadOnly || v.Mode == "ro" // TODO other modes?
}
