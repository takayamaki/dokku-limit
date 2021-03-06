package resource

import (
	"fmt"
	units "github.com/docker/go-units"
	"github.com/dokku/dokku/plugins/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
)

type Type string

const (
	TypeMemory Type = "memory"
	TypeCPU    Type = "cpu"
)

var ResourceTypes = []Type{
	TypeMemory,
	TypeCPU,
}

type Resources map[Type]int64

// Returns formatted docker arguments
func (r Resources) DockerOptions() []string {
	args := make([]string, len(r))

	for typ, limit := range r {
		args = append(args, FormatDocker(typ, limit))
	}

	return args
}

func Defaults() Resources {
	return LoadDefaults()
}

func SetDefaults(r Resources) {
	defaults := Defaults()
	if r == nil {
		r = defaults
	}
	for typ, limit := range defaults {
		if _, ok := r[typ]; !ok  {
			r[typ] = limit
		}
	}
}

// return starting defaults
func SystemDefaults() Resources {
	return Resources{
		TypeMemory: 1073741824,	// 1GB
		TypeCPU: int64(100),
	}
}

type Limits map[string]Resources

// Save limits
func (l Limits) SaveToApp(appName string) error {
	cleanLimits(l)
	filePath := LimitFilePath(appName)
	raw, _ := yaml.Marshal(l)
	err := ioutil.WriteFile(filePath, raw, 0644)
	return err
}

func Format(typ Type, limit int64) string {
	switch typ {
	case TypeMemory:
		return units.BytesSize(float64(limit))
	case TypeCPU:
		return fmt.Sprintf("%d%%", limit)
	default:
		return strconv.FormatInt(limit, 10)
	}
}

func FormatDocker(typ Type, limit int64) string {
	switch typ {
	case TypeMemory:
		return fmt.Sprintf("--memory=%d", limit)
	case TypeCPU:
		numCPU := runtime.NumCPU()
		cpus := (float64(numCPU) / float64(100) * float64(limit))
		return fmt.Sprintf("--cpus=\"%.2g\"", cpus)
	}
	return ""
}

func ToType(s string) (Type, bool) {
	switch s {
	case string(TypeMemory):
		return TypeMemory, true
	case string(TypeCPU):
		return TypeCPU, true
	default:
		return Type(""), false
	}
}

func Parse(limits []string) Resources {
	resources := Resources{}
	for _, limit := range limits {
		typVal := strings.SplitN(limit, "=", 2)
		if len(typVal) != 2 {
			common.LogFail(fmt.Sprintf("invalid resource limit: %q", limit))
		}
		typ, ok := ToType(typVal[0])
		if !ok {
			common.LogFail(fmt.Sprintf("invalid resource limit type: %q", typVal))
		}
		val, err := ParseLimit(typ, typVal[1])
		if err != nil {
			common.LogFail(fmt.Sprintf("invalid resource limit value: %q", typVal[1]))
		}
		resources[typ] = val
	}
	return resources
}

func ParseLimit(typ Type, s string) (int64, error) {
	switch typ {
	case TypeMemory:
		return units.RAMInBytes(s)
	case TypeCPU:
		val, err := units.FromHumanSize(s)
		if val > 100 || val <= 0 {
			return -1, fmt.Errorf("Invalid CPU value, should be between 1 - 100")
		}
		return val, err
	default:
		return units.FromHumanSize(s)
	}
}

func AppRoot(appName string) (appRoot string) {
	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	appRoot = strings.Join([]string{dokkuRoot, appName}, "/")
	return appRoot
}

func LimitFilePath(appName string) (filePath string) {
	appRoot := AppRoot(appName)
	return strings.Join([]string{appRoot, "RESOURCES.yml"}, "/")
}

func DefaultsFilePath() (filePath string) {
	libroot := common.MustGetEnv("DOKKU_LIB_ROOT")
	return strings.Join([]string{libroot, "data", "limit", "RESOURCES.yml"}, "/")
}

// load limits from resource app file
func LoadForApp(appName string) Limits {
	filePath := LimitFilePath(appName)

	if !common.FileExists(filePath) {
		return Limits{}
	}

	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		common.LogFail(err.Error())
	}

	limits := Limits{}
	err = yaml.Unmarshal(raw, &limits)

	cleanLimits(limits)

	return limits
}

// load default resource limits
func LoadDefaults() Resources {
	filePath := DefaultsFilePath()
	resources := Resources{}

	if !common.FileExists(filePath) {
		return SystemDefaults()
	}

	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		common.LogFail(err.Error())
	}

	err = yaml.Unmarshal(raw, &resources)

	return resources
}

// Save default resource limits
func SaveDefaults(r Resources) error {
	filePath := DefaultsFilePath()
	raw, _ := yaml.Marshal(r)
	err := ioutil.WriteFile(filePath, raw, 0644)
	return err
}

// Get the processes for an app
func GetAppProcs(appName string) map[string]bool {
	appRoot := AppRoot(appName)
	filePath := strings.Join([]string{appRoot, "DOKKU_SCALE"}, "/")
	procs := make(map[string]bool)

	if !common.FileExists(filePath) {
		return procs
	}

	lines, err := common.FileToSlice(filePath)
	if err != nil {
		common.LogFail(err.Error())
	}

	for _, line := range lines {
		procs[strings.Split(line, "=")[0]] = true
	}

	return procs
}

func cleanLimits(l Limits) {
	for procName, resources := range l {
		if len(resources) == 0 {
			delete(l, procName)
		}
	}
}
