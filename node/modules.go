package node

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type ModuleRunner interface {
	Run(context.Context, *Node) error
}

func (node *Node) runModules(ctx context.Context, modules []ModuleRunner) {
	names := make([]string, 0, len(modules))
	for _, mod := range modules {
		var name string
		if str, ok := mod.(fmt.Stringer); ok {
			name = str.String()
		} else {
			name = reflect.TypeOf(mod).Name()
		}
		names = append(names, name)
		mod := mod
		go func() {
			err := mod.Run(ctx, node)
			if err != nil {
				log.Printf("module %v error: %v\n", name, err)
			}
		}()
	}
	log.Println("modules:", strings.Join(names, " "))
}
