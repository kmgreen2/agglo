package core

import (
	"fmt"
	"reflect"
)

type mapNode interface {
	Value() interface{}
}

type DictNode struct {
	value map[string]interface{}
}

func (node DictNode) Value() interface{} {
	return node.value
}

func (node *DictNode) Set(key string, val mapNode) {
	node.value[key] = val
}

type ListNode struct {
	value []interface{}
}

func (node ListNode) Value() interface{} {
	return node.value
}

func (node *ListNode) Append(val mapNode) {
	node.value = append(node.value, val)
}

type NumericNode struct {
	value float64
}

func (node NumericNode) Value() interface{} {
	return node.value
}

func (node *NumericNode) Set(val float64) {
	node.value = val
}

type StringNode struct {
	value string
}

func (node StringNode) Value() interface{} {
	return node.value
}

func (node *StringNode) Set(val string) {
	node.value = val
}

type BoolNode struct {
	value bool
}

func (node BoolNode) Value() interface{} {
	return node.value
}

func (node *BoolNode) Set(val bool) {
	node.value = val
}

type Map struct {
	root *DictNode
}

func mapToString(node mapNode, leading string) string {
	switch v := node.(type) {
	case *DictNode:
		s := "{\n"
		internalDict := v.Value().(map[string]interface{})
		i := 0
		for k, v := range internalDict {
			s += leading + fmt.Sprintf("'%s': ", k) + mapToString(v.(mapNode), leading+"\t")
			if i < len(internalDict) - 1 || len(internalDict) == 1 {
				s += "\n"
			}
		}
		s += leading + "}"
		return s
	case *ListNode:
		s := "[\n"
		internalList := v.Value().([]interface{})
		for _, v := range internalList {
			s += leading + mapToString(v.(mapNode), leading+"\t") + "\n"
		}
		s += leading + "]"
		return s
	case *NumericNode:
		return fmt.Sprintf("%.2f", v.Value().(float64))
	case *StringNode:
		return v.Value().(string)
	case *BoolNode:
		return fmt.Sprintf("%s", v.Value().(bool))
	}
	return ""
}

func (m *Map) String() string {
	return mapToString(m.root, "")
}

func buildMap(inNode mapNode, in interface{}) error {
	switch node := inNode.(type) {
	case *DictNode:
		if _, ok := in.(map[string]interface{}); !ok {
			return fmt.Errorf("")
		}
		elements := in.(map[string]interface{})
		for k, v := range elements {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				dictNode := &DictNode{make(map[string]interface{})}
				node.Set(k, dictNode)
				err := buildMap(dictNode, v.(map[string]interface{}))
				if err != nil {
					return err
				}
			} else if reflect.TypeOf(v).Kind() == reflect.Slice {
				listNode := &ListNode{}
				node.Set(k, listNode)
				err := buildMap(listNode, v.([]interface{}))
				if err != nil {
					return err
				}
			} else if reflect.TypeOf(v).Kind() == reflect.String {
				node.Set(k, &StringNode{v.(string)})
			} else if reflect.TypeOf(v).Kind() == reflect.Bool {
				node.Set(k, &BoolNode{v.(bool)})
			} else if v, err := GetNumeric(v); err == nil {
				node.Set(k, &NumericNode{v})
			}
		}
	case *ListNode:
		if _, ok := in.([]interface{}); !ok {
			return fmt.Errorf("")
		}
		elements := in.([]interface{})
		for _, v := range elements {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				dictNode := &DictNode{make(map[string]interface{})}
				node.Append(dictNode)
				err := buildMap(dictNode, v.(map[string]interface{}))
				if err != nil {
					return err
				}
			} else if reflect.TypeOf(v).Kind() == reflect.Slice {
				listNode := &ListNode{}
				node.Append(listNode)
				err := buildMap(listNode, v.([]interface{}))
				if err != nil {
					return err
				}
			} else if reflect.TypeOf(v).Kind() == reflect.String {
				node.Append(&StringNode{v.(string)})
			} else if reflect.TypeOf(v).Kind() == reflect.Bool {
				node.Append(&BoolNode{v.(bool)})
			} else if v, err := GetNumeric(v); err == nil {
				node.Append(&NumericNode{v})
			}
		}
	}
	return nil
}

func NewMap(in map[string]interface{}) (*Map, error) {
	newMap := &Map {
		&DictNode{
			value: make(map[string]interface{}),
		},
	}
	err := buildMap(newMap.root, in)
	if err != nil {
		return nil, err
	}
	return newMap, nil
}

