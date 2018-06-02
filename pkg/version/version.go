package version

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Version struct {
	Tag       string
	BuildArch string
	GitCommit string
	BuildDate time.Time
}

func (v *Version) StringForCommandLine() string {
	return strings.TrimSpace(fmt.Sprintf(`
Podlike (https://github.com/rycus86/podlike)
--------------------------------------------
Version    : %s-%s
Git commit : %s
Built at   : %s
`, v.Tag, v.BuildArch, v.GitCommit, v.BuildDate.Format(time.RFC3339)))
}

func Parse() *Version {
	timestamp := getEnv("BUILD_TIMESTAMP", "0")
	timeAsInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		timeAsInt = 0
	}

	return &Version{
		Tag:       getEnv("VERSION", "dev"),
		BuildArch: getEnv("BUILD_ARCH", "unknown"),
		GitCommit: getEnv("GIT_COMMIT", "unknown"),
		BuildDate: time.Unix(timeAsInt, 0),
	}
}

func getEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return defaultValue
	}
}
