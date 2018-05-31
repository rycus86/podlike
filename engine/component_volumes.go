package engine

import (
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-units"
	"github.com/mitchellh/mapstructure"
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
			Type:     v.getMountType(),
			Source:   v.Source,
			Target:   v.Target,
			ReadOnly: v.isReadOnly(),
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

func (c *Component) parseVolumes() ([]*Volume, error) {
	if len(c.Volumes) == 0 {
		return []*Volume{}, nil
	}

	converted := make([]*Volume, len(c.Volumes), len(c.Volumes))

	for idx, item := range c.Volumes {
		if asString, ok := item.(string); ok {
			volume := Volume{}
			parts := strings.Split(asString, ":")

			switch len(parts) {
			case 1:
				volume.Target = parts[0]
			case 2:
				volume.Source = parts[0]
				volume.Target = parts[1]
			case 3:
				volume.Source = parts[0]
				volume.Target = parts[1]
				volume.Mode = parts[2]
			}

			converted[idx] = &volume
		} else {
			var volume Volume

			err := mapstructure.Decode(item, &volume)
			if err != nil {
				return nil, err
			}

			converted[idx] = &volume
		}
	}

	return converted, nil
}
