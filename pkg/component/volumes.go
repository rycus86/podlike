package component

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-units"
	"github.com/mitchellh/mapstructure"
	"github.com/rycus86/podlike/pkg/volume"
	"strings"
)

func (c *Component) getMounts() ([]mount.Mount, error) {
	volumes, err := c.parseVolumes()
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		return []mount.Mount{}, nil
	}

	mounts := make([]mount.Mount, 0, len(volumes))

	for _, v := range volumes {
		mnt := mount.Mount{
			Type:     v.GetMountType(),
			Source:   v.Source,
			Target:   v.Target,
			ReadOnly: v.IsReadOnly(),
		}

		sharedVolumeSource := c.client.GetSharedVolumeSource(mnt.Source)
		if sharedVolumeSource != "" {
			mnt.Source = sharedVolumeSource
		}

		if v.Bind.Propagation != "" {
			mnt.BindOptions = &mount.BindOptions{
				Propagation: mount.Propagation(v.Bind.Propagation),
			}
		}

		if v.Volume.NoCopy {
			mnt.VolumeOptions = &mount.VolumeOptions{
				NoCopy: true,
			}
		}

		if v.Tmpfs.Size != "" {
			size, err := units.FromHumanSize(v.Tmpfs.Size)
			if err != nil {
				return nil, err
			}

			mnt.TmpfsOptions = &mount.TmpfsOptions{
				SizeBytes: size,
			}
		}

		mounts = append(mounts, mnt)
	}

	return mounts, nil
}

func (c *Component) parseVolumes() ([]*volume.Volume, error) {
	if len(c.Volumes) == 0 {
		return []*volume.Volume{}, nil
	}

	converted := make([]*volume.Volume, len(c.Volumes), len(c.Volumes))

	for idx, item := range c.Volumes {
		if asString, ok := item.(string); ok {
			v := volume.Volume{}
			parts := strings.Split(asString, ":")

			switch len(parts) {
			case 1:
				v.Target = parts[0]
			case 2:
				v.Source = parts[0]
				v.Target = parts[1]
			case 3:
				v.Source = parts[0]
				v.Target = parts[1]
				v.Mode = parts[2]
			}

			converted[idx] = &v
		} else {
			var v volume.Volume

			err := mapstructure.Decode(item, &v)
			if err != nil {
				return nil, err
			}

			converted[idx] = &v
		}
	}

	return converted, nil
}
