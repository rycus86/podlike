package engine

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func (c *Component) copyFilesIfNecessary() error {
	for key, value := range c.client.container.Config.Labels {
		if strings.Index(key, "pod.copy.") >= 0 {
			if target := strings.TrimPrefix(key, "pod.copy."); target != c.Name {
				continue
			}

			configs, err := parseCopyConfig(value)
			if err != nil {
				return err
			}

			for _, config := range configs {
				targetDir, targetFilename := path.Split(config.Target)
				reader, err := createTar(config.Source, targetFilename)
				if err != nil {
					return err
				}

				fmt.Println("Copying", config.Source, "to", c.Name, "@", config.Target, "...")

				err = c.client.api.CopyToContainer(
					context.TODO(), c.container.ID, targetDir, reader, types.CopyToContainerOptions{})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func parseCopyConfig(definition string) ([]CopyConfig, error) {
	var value interface{}

	if err := yaml.Unmarshal([]byte(definition), &value); err != nil {
		return nil, err
	}

	// a simple string, like:
	//  pod.copy.sample: /source:/target
	if strValue, ok := value.(string); ok {
		parts := strings.Split(strValue, ":")
		if len(parts) != 2 {
			return nil, errors.New(fmt.Sprintf("invalid pod.copy configuration: %s", value))
		}

		return []CopyConfig{
			{
				Source: strings.Trim(parts[0], " \t\n"),
				Target: strings.Trim(parts[1], " \t\n"),
			},
		}, nil
	}

	// a mapping, like:
	//  pod.copy.sample: |
	//    /source: /target
	if _, isMap := value.(map[interface{}]interface{}); isMap {
		mapped, err := asStringToStringMap(value)
		if err != nil {
			return nil, err
		}

		parsed := make([]CopyConfig, 0, len(mapped))

		for source, target := range mapped {
			source := strings.Trim(source, " \t\n")
			target := strings.Trim(target, " \t\n")

			if source == "" || target == "" {
				return nil, errors.New(fmt.Sprintf(
					"invalid pod.copy configuration: %s [%s:%s]", value, source, target))
			}

			parsed = append(parsed, CopyConfig{Source: source, Target: target})
		}

		return parsed, nil
	}

	// a sequence, like:
	//  pod.copy.sample: |
	//    - /source:/target
	if items, err := asStringSlice(value); err == nil {
		parsed := make([]CopyConfig, 0, len(items))

		for _, item := range items {
			parts := strings.Split(item, ":")

			if len(parts) != 2 {
				return nil, errors.New(fmt.Sprintf(
					"invalid pod.copy configuration: %s [%s]", value, item))
			}

			source := strings.Trim(parts[0], " \t\n")
			target := strings.Trim(parts[1], " \t\n")

			if source == "" || target == "" {
				return nil, errors.New(fmt.Sprintf(
					"invalid pod.copy configuration: %s [%s:%s]", value, source, target))
			}

			parsed = append(parsed, CopyConfig{Source: source, Target: target})
		}

		return parsed, nil
	} else {
		return nil, err
	}

	return nil, errors.New(fmt.Sprintf("invalid pod.copy configuration: %s (%T)", value, value))
}

func createTar(path, filename string) (io.Reader, error) {
	var b bytes.Buffer

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tw := tar.NewWriter(&b)
	hdr := tar.Header{
		Name: filename,
		Mode: 0644,
		Size: fi.Size(),
	}
	if err := tw.WriteHeader(&hdr); err != nil {
		return nil, err
	}

	if _, err = tw.Write(contents); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &b, nil
}
