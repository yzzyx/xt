package main

import (
	"fmt"
	"strings"
)

func main() {
	tmpl := `<html><head>{% block test %}blah{% block x2 %}in x2{% endblock %}{% endblock %}xoxoxo`
	tmpl += `{% if abc %}xx{% else %}yy{% endif %}</html>`

	t := NewTree("test")
	err := t.Parse(tmpl)
	if err != nil {
		fmt.Println("err:", err)
	}

	var fn func(indent int) Walker
	fn = func(indent int) Walker {
		return func(node Node) Walker {
			fmt.Printf("%s%+v\n", strings.Repeat("\t", indent), node)
			return fn(indent + 1)
		}
	}

	err = t.Walk(fn(0))
	if err != nil {
		fmt.Println("cannot walk:", err)
	}
}
