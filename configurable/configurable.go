package configurable

import (
	`encoding/json`
	`errors`
	`flag`
	`os`
	`reflect`
	`time`

	`github.com/go-ini/ini`
	`gopkg.in/yaml.v3`
)

type IConfigurable interface {
	Int(name string, value int, usage string) *int
	Int64(name string, value int64, usage string) *int64
	Float(name string, value float64, usage string) *float64
	Duration(name string, value time.Duration, usage string) *time.Duration
	String(name string, value string, usage string) *string
	Bool(name string, value bool, usage string) *bool
	Parse(filename string) error
	Value(name string) interface{}
}
type Configurable struct {
	flags map[string]interface{}
}

func NewConfigurable() IConfigurable {
	return &Configurable{flags: make(map[string]interface{})}
}

func (c *Configurable) Int(name string, value int, usage string) *int {
	var i = flag.Int(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) Int64(name string, value int64, usage string) *int64 {
	var i = flag.Int64(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) Float(name string, value float64, usage string) *float64 {
	var i = flag.Float64(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) Duration(name string, value time.Duration, usage string) *time.Duration {
	var i = flag.Duration(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) String(name string, value string, usage string) *string {
	var s = flag.String(name, value, usage)
	c.flags[name] = s
	return s
}

func (c *Configurable) Bool(name string, value bool, usage string) *bool {
	var b = flag.Bool(name, value, usage)
	c.flags[name] = b
	return b
}

func (c *Configurable) Parse(filename string) error {
	flag.Parse()
	err := c.overrideWithFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configurable) Value(name string) interface{} {
	return c.flags[name]
}

func (c *Configurable) overrideWithFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	ext := reflect.ValueOf(filename).Type().String()
	switch ext {
	case ".json":
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			return err
		}
		for key, value := range jsonData {
			if c.flags[key] != nil {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(value))
			}
		}
	case ".yaml", ".yml":
		var yamlData map[string]interface{}
		err = yaml.Unmarshal(data, &yamlData)
		if err != nil {
			return err
		}
		for key, value := range yamlData {
			if c.flags[key] != nil {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(value))
			}
		}
	case ".ini":
		cfg, err := ini.Load(data)
		if err != nil {
			return err
		}
		for key, _ := range c.flags {
			if cfg.Section("").HasKey(key) {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(cfg.Section("").Key(key).Value()))
			}
		}
	default:
		return errors.New("unknown file type")
	}
	return nil
}
