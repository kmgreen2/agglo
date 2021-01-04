package main

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kmgreen2/agglo/generated/proto"
	"math/rand"
	"strings"
)

type ReferenceKey struct {
	schemaName string
	path string
}

type DependencyGraph struct {
	vertices map[string]bool
	forwardEdges map[string][]string
	backwardEdges map[string][]string
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph {
		make(map[string]bool),
		make(map[string][]string),
		make(map[string][]string),
	}
}

func (g *DependencyGraph) AddEdge(from, to string) {
	g.vertices[from] = true
	g.vertices[to] = true
	g.forwardEdges[from] = append(g.forwardEdges[from], to)
	g.backwardEdges[to] = append(g.backwardEdges[to], from)
}

func (g *DependencyGraph) Length() int {
	return len(g.vertices)
}

func (g *DependencyGraph) OrphanedVertex() (string, error) {
	for v, _ := range g.vertices {
		if _, ok := g.backwardEdges[v]; !ok {
			return v, nil
		}
		if len(g.backwardEdges) == 0 {
			return v, nil
		}
	}
	return "", fmt.Errorf("Cannot find Orphaned vertex")
}

func (g *DependencyGraph) Delete(v string) error {
	delete(g.vertices, v)

	if edges, ok := g.forwardEdges[v]; ok {
		for _, to := range edges {
			numEdges := len(g.backwardEdges[to])
			for i := 0; i < numEdges; i++ {
				if strings.Compare(g.backwardEdges[to][i], v) == 0 {
					g.backwardEdges[to] = append(g.backwardEdges[to][:i], g.backwardEdges[to][i+1:]...)
					break
				}
			}
		}
	}
	g.forwardEdges[v] = []string{}
	return nil
}

func (g *DependencyGraph) Copy() *DependencyGraph {
	vertices := make(map[string]bool)
	forwardEdges := make(map[string][]string)
	backwardEdges := make(map[string][]string)

	for k, v := range g.vertices {
		vertices[k] = v
	}

	for k, _ := range g.forwardEdges {
		for _, v := range g.forwardEdges[k] {
			forwardEdges[k] = append(forwardEdges[k], v)
		}
	}

	for k, _ := range g.backwardEdges {
		for _, v := range g.backwardEdges[k] {
			backwardEdges[k] = append(backwardEdges[k], v)
		}
	}

	return &DependencyGraph{
		vertices: vertices,
		forwardEdges: forwardEdges,
		backwardEdges: backwardEdges,
	}
}

func TopSort(graph *DependencyGraph) ([]string, error) {
	var sorted  []string
	g := graph.Copy()
	// While there exists a vertex with no inbound edges
	for {
		if g.Length() == 0 {
			return sorted, nil
		}
		// Choose a vertex with no inbound edges (no backwards entries)
		v, err := g.OrphanedVertex()
		if err != nil {
			return nil, err
		}
		// Remove vertex and its dependent edges
		err = g.Delete(v)
		if err != nil {
			return nil, err
		}
	}
}

type GeneratorState struct {
	schemas api.Schemas
	references map[ReferenceKey]int
	referenceGraph *DependencyGraph
}

func (g *GeneratorState) drawSchema() *api.Schema {
	schemaDraw := rand.Float64()

	curr := float64(0)
	for i, v := range g.schemas.SchemaDistribution {
		if schemaDraw >= curr && schemaDraw <= v {
			return g.schemas.Schemas[i]
		}
		curr = v
	}
	return nil
}

func (g *GeneratorState) generateValue(value *api.Value) (interface{}, error) {
	var err error
	switch val := value.Values.(type) {
	case *api.Value_RandomString:
		s := ""
		length := int32(rand.Float64() * float64(val.RandomString.MaxLen - val.RandomString.MinLen)) + val.RandomString.MinLen
		for i := int32(0); i < length; i++ {
			idx := rand.Int() % len(val.RandomString.Alphabet)
			s += string(val.RandomString.Alphabet[idx])
		}
		return s, nil
	case *api.Value_VocabString:
		idx := rand.Int() % len(val.VocabString.Vocab)
		return val.VocabString.Vocab[idx], nil
	case *api.Value_FixedString:
		return val.FixedString.Value, nil
	case *api.Value_RandomNumeric:
		delta := val.RandomNumeric.Max - val.RandomNumeric.Min
		return val.RandomNumeric.Min + (rand.Float64() * delta), nil
	case *api.Value_NumericSet:
		idx := rand.Int() % len(val.NumericSet.Values)
		return val.NumericSet.Values[idx], nil
	case *api.Value_FixedNumeric:
		return val.FixedNumeric.Value, nil
	case *api.Value_Boolean:
		if rand.Float64() >= float64(0.5) {
			return true, nil
		} else {
			return false, nil
		}
	case *api.Value_Dict:
		out := make(map[string]interface{})
		for k, v := range val.Dict.Kvs {
			out[k], err = g.generateValue(v)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case *api.Value_List:
		length := int32(rand.Float64() * float64(val.List.MaxLen - val.List.MinLen)) + val.List.MinLen
		slice := make([]interface{}, length)
		for i := int32(0); i < length; i++ {
			slice[i], err = g.generateValue(val.List.Value)
			if err != nil {
				return nil, err
			}
		}
	case *api.Value_Reference:
	}
	return nil, nil
}

func (g *GeneratorState) Generate() (map[string]interface{}, error) {
	var err error
	out := make(map[string]interface{})

	// Use dependency graph to determine if there is enough budget for references.  If not, generate
	// a schema that will free up budget
	schema := g.drawSchema()

	for k, v := range schema.Root.Kvs {
		out[k], err = g.generateValue(v)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (g *GeneratorState) getDependencyGraph(thisSchema string, value *api.Value) error {
	var err error
	switch val := value.Values.(type) {
	case *api.Value_Dict:
		for _, v := range val.Dict.Kvs {
			err = g.getDependencyGraph(thisSchema, v)
			if err != nil {
				return err
			}
		}
	case *api.Value_List:
		return g.getDependencyGraph(thisSchema, val.List.Value)
	case *api.Value_Reference:
		g.referenceGraph.AddEdge(val.Reference.SchemaName, thisSchema)
	}
	return nil
}

func (g *GeneratorState) validateDependencies() error {
	_, err := TopSort(g.referenceGraph)
	if err != nil {
		return err
	}
	return nil
}

func SchemasFromJson(schemasJson []byte) (*api.Schemas, error) {
	var schemasPb api.Schemas
	byteBuffer := bytes.NewBuffer(schemasJson)
	err := jsonpb.Unmarshal(byteBuffer, &schemasPb)
	if err != nil {
		return nil, err
	}
	return &schemasPb, nil
}

func main() {

}
