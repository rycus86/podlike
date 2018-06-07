package template

import (
	"testing"
)

func TestMerge_Maps(t *testing.T) {
	target := map[string]interface{}{}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"image": "example/image",
			"labels": map[string]interface{}{
				"label.source": "m1",
			},
		},
	}

	m2 := map[string]interface{}{
		"app": map[string]interface{}{
			"labels": map[string]interface{}{
				"label.source":     "ignored",
				"additional.label": "m2",
			},
			"read_only": true,
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if image, ok := mApp["image"]; !ok || image != "example/image" {
		t.Error("missing image key")
	} else if labels, ok := mApp["labels"]; !ok {
		t.Fatal("missing labels")
	} else if mLabels, ok := labels.(map[string]interface{}); !ok {
		t.Fatal("invalid labels")
	} else if labelSource, ok := mLabels["label.source"]; !ok || labelSource != "m1" {
		t.Error("missing or invalid label value")
	}

	mergeRecursively(target, m2)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if readOnly, ok := mApp["read_only"]; !ok || readOnly != true {
		t.Error("missing or invalid read_only key")
	} else if labels, ok := mApp["labels"]; !ok {
		t.Fatal("missing labels")
	} else if mLabels, ok := labels.(map[string]interface{}); !ok {
		t.Fatal("invalid labels")
	} else if labelSource, ok := mLabels["label.source"]; !ok || labelSource != "m1" {
		t.Error("missing or invalid label value")
	} else if aLabel, ok := mLabels["additional.label"]; !ok || aLabel != "m2" {
		t.Error("missing or invalid label value")
	}
}

func TestMerge_WithSlices(t *testing.T) {
	target := map[string]interface{}{
		"app": map[string]interface{}{
			"volumes": []interface{}{
				"v1:/var/v1",
				"v2:/var/v2",
			},
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"volumes": []interface{}{
				"v3:/var/v3",
			},
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if volumes, ok := mApp["volumes"]; !ok {
		t.Fatal("missing volumes")
	} else if !sliceMatches(volumes, "v1:/var/v1", "v2:/var/v2", "v3:/var/v3") {
		t.Fatal("invalid volumes")
	}
}

func TestMerge_MixedMapsAndSlices(t *testing.T) {
	target := map[string]interface{}{
		"app": map[string]interface{}{
			"labels": map[string]interface{}{
				"label1": "first",
				"label2": "second",
			},
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"labels": []interface{}{
				"label2=ignored",
				"label3=third",
			},
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if labels, ok := mApp["labels"]; !ok {
		t.Fatal("missing labels")
	} else if mLabels, ok := labels.(map[string]interface{}); !ok {
		t.Fatal("invalid labels")
	} else if label1, ok := mLabels["label1"]; !ok || label1 != "first" {
		t.Error("missing or invalid label value")
	} else if label2, ok := mLabels["label2"]; !ok || label2 != "second" {
		t.Error("missing or invalid label value")
	} else if label3, ok := mLabels["label3"]; !ok || label3 != "third" {
		t.Error("missing or invalid label value")
	}
}

func TestMerge_MixedSlicesAndMaps(t *testing.T) {
	target := map[string]interface{}{
		"app": map[string]interface{}{
			"labels": []interface{}{
				"label1=first",
				"label2=second",
			},
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"labels": map[string]interface{}{
				"label2": "ignored",
				"label3": "third",
			},
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if labels, ok := mApp["labels"]; !ok {
		t.Fatal("missing labels")
	} else if mLabels, ok := labels.(map[string]interface{}); !ok {
		t.Fatal("invalid labels")
	} else if label1, ok := mLabels["label1"]; !ok || label1 != "first" {
		t.Error("missing or invalid label value")
	} else if label2, ok := mLabels["label2"]; !ok || label2 != "second" {
		t.Error("missing or invalid label value")
	} else if label3, ok := mLabels["label3"]; !ok || label3 != "third" {
		t.Error("missing or invalid label value")
	}
}

func TestMerge_Deeper(t *testing.T) {
	target := map[string]interface{}{
		"app": map[string]interface{}{
			"deploy": map[string]interface{}{
				"mode": "global",
			},
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"deploy": map[string]interface{}{
				"placement": map[string]interface{}{
					"constraints": []interface{}{
						"node.role == manager",
					},
				},
				"resources": map[string]interface{}{
					"limits": map[string]interface{}{
						"memory": "50m",
					},
				},
			},
		},
	}

	m2 := map[string]interface{}{
		"app": map[string]interface{}{
			"deploy": map[string]interface{}{
				"placement": map[string]interface{}{
					"constraints": []interface{}{
						"node.labels.test == xyz",
					},
				},
				"resources": map[string]interface{}{
					"limits": map[string]interface{}{
						"cpu": 0.2,
					},
				},
			},
		},
	}

	mergeRecursively(target, m1)
	mergeRecursively(target, m2)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if deploy, ok := mApp["deploy"]; !ok {
		t.Fatal("missing deploy key")
	} else if mDeploy, ok := deploy.(map[string]interface{}); !ok {
		t.Fatal("invalid deploy key")
	} else if mode, ok := mDeploy["mode"]; !ok || mode != "global" {
		t.Error("missing or invalid mode")
	} else if placement, ok := mDeploy["placement"]; !ok {
		t.Fatal("missing placement key")
	} else if mPlacement, ok := placement.(map[string]interface{}); !ok {
		t.Fatal("invalid placement key")
	} else if constraints, ok := mPlacement["constraints"]; !ok {
		t.Fatal("missing constraints key")
	} else if !sliceMatches(constraints,
		"node.role == manager",
		"node.labels.test == xyz",
	) {
		t.Error("invalid constraints")
	} else if resources, ok := mDeploy["resources"]; !ok {
		t.Fatal("missing resources key")
	} else if mResources, ok := resources.(map[string]interface{}); !ok {
		t.Fatal("invalid resources key")
	} else if limits, ok := mResources["limits"]; !ok {
		t.Fatal("missing limits key")
	} else if mLimits, ok := limits.(map[string]interface{}); !ok {
		t.Fatal("invalid limits key")
	} else if memory, ok := mLimits["memory"]; !ok || memory != "50m" {
		t.Error("invalid memory limit")
	} else if cpu, ok := mLimits["cpu"]; !ok || cpu != 0.2 {
		t.Error("invalid cpu limit")
	}
}

func TestMerge_StringIntoSlice(t *testing.T) {
	target := map[string]interface{}{
		"app": map[string]interface{}{
			"command": []interface{}{
				"ls", "-l",
			},
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"command": "-h",
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if command, ok := mApp["command"]; !ok {
		t.Fatal("missing command key")
	} else if !sliceMatches(command, "ls", "-l", "-h") {
		t.Error("invalid command")
	}
}

func TestMerge_SliceIntoStringIsNotSupported(t *testing.T) {
	// Hint: make sure it's a slice in the first place

	target := map[string]interface{}{
		"app": map[string]interface{}{
			"command": "/bin/app",
		},
	}

	m1 := map[string]interface{}{
		"app": map[string]interface{}{
			"command": []interface{}{
				"--this", "--gets", "--ignored",
			},
		},
	}

	mergeRecursively(target, m1)

	if app, ok := target["app"]; !ok {
		t.Fatal("missing root element")
	} else if mApp, ok := app.(map[string]interface{}); !ok {
		t.Fatal("invalid root element")
	} else if command, ok := mApp["command"]; !ok || command != "/bin/app" {
		t.Fatal("missing or invalid command")
	}
}

func sliceMatches(actual interface{}, expected ...interface{}) bool {
	if s, ok := actual.([]string); ok {
		// convert []string slices to []interface{}
		var i []interface{}

		for _, item := range s {
			i = append(i, item)
		}

		return sliceMatches(i, expected...)

	} else if c, ok := actual.([]interface{}); !ok {
		return false
	} else if len(c) != len(expected) {
		return false
	} else {
		for idx, v := range c {
			if expected[idx] != v {
				return false
			}
		}
	}

	return true
}
