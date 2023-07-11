package deepcopy

import (
	"reflect"
	"strings"
)

type fieldDetail struct {
	field    *reflect.StructField
	key      string
	ignored  bool
	required bool
}

func parseTag(detail *fieldDetail) {
	tagValue, ok := detail.field.Tag.Lookup(DefaultTagName)
	detail.key = detail.field.Name
	if !ok {
		return
	}

	tags := strings.Split(tagValue, ",")
	switch {
	case tags[0] == "-":
		detail.ignored = true
	case tags[0] != "":
		detail.key = tags[0]
	}

	for _, tagOpt := range tags[1:] {
		if tagOpt == "required" {
			detail.required = true
		}
	}
}
