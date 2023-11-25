package embed

import "github.com/zekrotja/ken"

var Cmds map[string]interface{}

func init() {
	Cmds = make(map[string]interface{})
}

type Cmd struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func List() []Cmd {
	var result []Cmd
	for name, v := range Cmds {
		c, ok := v.(ken.Command)
		if ok {
			result = append(result, Cmd{
				Name:        name,
				Description: c.Description(),
			})
		}
	}
	return result
}
