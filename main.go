package main

import (
	"context"
	"fmt"
	"strings"
)

func main() {
	//tmpl := `<html><head>{% block test %}blah{% block x2 %}in x2{% endblock %}{% endblock %}xoxoxo`
	//tmpl += `{% if "abc" == somevar %}xx{% else %}yy{% endif %}</html>`
	tmpl := `{% bleep %}in bleep{% endbleep %}`

	t := NewTree("test")
	//t.RegisterTag("bleep", &TagBleep{})
	err := t.Parse(tmpl)
	if err != nil {
		fmt.Println("err:", err)
	}

	ctx := context.Background()
	var fn func(indent int) Walker
	fn = func(indent int) Walker {
		return func(node Node) Walker {
			v, err := node.Execute(ctx)
			if err != nil {
				v = err.Error()
			}
			fmt.Printf("%s%T %+v : %s\n", strings.Repeat("\t", indent), node, node, v)
			return fn(indent + 1)
		}
	}

	err = t.Walk(fn(0))
	if err != nil {
		fmt.Println("cannot walk:", err)
	}
}
