package steps

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseNodeValueDuration(value string) (time.Duration, error) {
	var s string
	if err := yaml.Unmarshal([]byte(value), &s); err != nil {
		return 0, errors.WithStack(err)
	}

	return util.ParseDuration(s)
}

func updateDesignMap(src []byte, dotPath string, newVal any) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(src, &root); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if len(root.Content) == 0 {
		return nil, fmt.Errorf("empty YAML document")
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("top-level must be a mapping node")
	}

	parts := strings.Split(dotPath, ".")
	cur := doc

	for i, p := range parts {
		last := i == len(parts)-1
		ki, val, found := findPair(cur, p)

		if !found {
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: p}
			if last {
				var nv *yaml.Node
				switch v := newVal.(type) {
				case *yaml.Node:
					nv = v
				default:
					var err error
					nv, err = coerceForExistingScalar(nil, newVal)
					if err != nil {
						return nil, fmt.Errorf("set %q: %w", p, err)
					}
				}
				cur.Content = append(cur.Content, keyNode, nv)
				break
			}

			mapNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			cur.Content = append(cur.Content, keyNode, mapNode)
			cur = mapNode

			continue
		}

		if last {
			nv, err := coerceForExistingScalar(val, newVal)
			if err != nil {
				return nil, fmt.Errorf("set %q: %w", p, err)
			}
			cur.Content[ki+1] = nv

			break
		}

		if val.Kind != yaml.MappingNode {
			val.Kind, val.Tag, val.Content = yaml.MappingNode, "!!map", nil
		}
		cur = val
	}

	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(&root); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	_ = enc.Close()

	return out.Bytes(), nil
}

func findPair(m *yaml.Node, key string) (int, *yaml.Node, bool) {
	if m.Kind != yaml.MappingNode {
		return -1, nil, false
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if k := m.Content[i]; k.Kind == yaml.ScalarNode && k.Value == key {
			return i, m.Content[i+1], true
		}
	}
	return -1, nil, false
}

func coerceForExistingScalar(existing *yaml.Node, v any) (*yaml.Node, error) {
	if existing == nil {
		switch val := v.(type) {
		case bool:
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(val)}, nil
		case int:
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(val)}, nil
		case int64:
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(val, 10)}, nil
		case float64:
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(val, 'f', -1, 64)}, nil
		default:
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprint(v)}, nil
		}
	}

	if existing.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("existing value is not a scalar (kind=%v)", existing.Kind)
	}

	switch existing.Tag {
	case "!!str":
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprint(v)}, nil
	case "!!int":
		iv, err := coerceInt(v)
		if err != nil {
			return nil, err
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(iv)}, nil
	case "!!bool":
		bv, err := coerceBool(v)
		if err != nil {
			return nil, err
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(bv)}, nil
	default:
		return nil, fmt.Errorf("unsupported scalar tag %q at value %q", existing.Tag, existing.Value)
	}
}

func coerceInt(v any) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case string:
		x, err := strconv.Atoi(t)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as int", t)
		}
		return x, nil
	default:
		return 0, fmt.Errorf("type %T cannot be coerced to int", v)
	}
}

func coerceBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		switch strings.ToLower(t) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse %q as bool", t)
		}
	default:
		return false, fmt.Errorf("type %T cannot be coerced to bool", v)
	}
}
