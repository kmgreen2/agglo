package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/util"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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

func (g *DependencyGraph) Orphaned(v string) bool {
	return g.backwardEdges[v] == nil || len(g.backwardEdges) == 0
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
		if len(g.backwardEdges[v]) == 0 {
			return v, nil
		}
	}
	return "", fmt.Errorf("Cannot find Orphaned vertex")
}

func (g *DependencyGraph) GetParents(v string) ([]string, error) {
	if backEdges, ok := g.backwardEdges[v]; ok {
		return backEdges, nil
	}
	return []string{}, nil
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

type ReferenceValue struct {
	value interface{}
	refCount int
}

type ReferenceValues struct {
	values []*ReferenceValue
}

func (r *ReferenceValues) Add(v interface{}, refCount int) {
	r.values = append(r.values, &ReferenceValue{v, refCount})
}

func (r *ReferenceValues) Num() int {
	return len(r.values)
}

func (r *ReferenceValues) Pop() (interface{}, error) {
	if r.Num() == 0 {
		return nil, util.NewInvalidError(fmt.Sprintf("no available references"))
	}
	val := r.values[0].value
	r.values[0].refCount--

	if r.values[0].refCount <= 0 {
		r.values = r.values[1:]
	}
	return val, nil
}

type SchemaReferences struct {
	pathValues map[string]*ReferenceValues
}

func (r *SchemaReferences) Add(path string, v interface{}, refCount int) {
	if _, ok := r.pathValues[path]; !ok {
		r.pathValues[path] = &ReferenceValues{}
	}
	r.pathValues[path].Add(v, refCount)
}

func (r *SchemaReferences) Pop(path string) (interface{}, error) {
	if _, ok := r.pathValues[path]; ok {
		return r.pathValues[path].Pop()
	}
	return nil, util.NewInvalidError(fmt.Sprintf("references do not exist for %s", path))
}

func (r *SchemaReferences) HasPath(path string) bool {
	if _, ok := r.pathValues[path]; ok {
		return r.pathValues[path].Num() > 0
	}
	return false
}

func (r *SchemaReferences) Num(path string) int {
	if _, ok := r.pathValues[path]; ok {
		return r.pathValues[path].Num()
	}
	return 0
}

func (r *SchemaReferences) Paths() []string {
	var paths []string
	if r.pathValues != nil {
		for k, _ := range r.pathValues {
			paths = append(paths, k)
		}
	}
	return paths
}

type GeneratorState struct {
	schemas *api.Schemas
	references map[string]*SchemaReferences
	referenceGraph *DependencyGraph
	sortedDependencies []string
	globalStrings map[string]string
	globalCounter map[string]int32
}

func NewGeneratorState(schemas *api.Schemas) (*GeneratorState, error) {
	generatorState := &GeneratorState{
		schemas: schemas,
		references: make(map[string]*SchemaReferences),
		referenceGraph: NewDependencyGraph(),
		globalStrings: make(map[string]string),
		globalCounter: make(map[string]int32),
	}

	err := generatorState.validateDependencies()
	if err != nil {
		return nil, err
	}
	return generatorState, nil
}

func (g *GeneratorState) drawSchema() *api.Schema {
	schemaDraw := rand.Float64()

	curr := float64(0)
	for i, v := range g.schemas.SchemaDistribution {
		if schemaDraw >= curr && schemaDraw < (v+curr) {
			return g.schemas.Schemas[i]
		}
		curr += v
	}
	return nil
}

func (g *GeneratorState) generateValue(value *api.Value, eventLocalstate map[string]string,
	eventGlobalState map[string]string) (interface{}, error) {
	var err error
	switch val := value.Values.(type) {
	case *api.Value_RandomString:
		var length int32
		var s string
		if len(val.RandomString.SharedName) > 0 {
			if _, ok := eventLocalstate[val.RandomString.SharedName]; ok {
				s = eventLocalstate[val.RandomString.SharedName]
			}
		}

		if len(val.RandomString.ReadStringState) > 0 {
			if state, ok := g.globalStrings[val.RandomString.ReadStringState]; ok {
				s = state
			}
		}

		if len(s) == 0 {
			length = int32(rand.Float64() * float64(val.RandomString.MaxLen - val.RandomString.MinLen)) + val.RandomString.MinLen
			for i := int32(0); i < length; i++ {
				idx := rand.Int() % len(val.RandomString.Alphabet)
				s += string(val.RandomString.Alphabet[idx])
			}
		}
		if val.RandomString.MaxLen == val.RandomString.MinLen {
			length = val.RandomString.MaxLen
		}

		if len(val.RandomString.SharedName) > 0 {
			eventLocalstate[val.RandomString.SharedName] = s
		}

		if len(val.RandomString.StoreStringState) > 0 {
			eventGlobalState[val.RandomString.StoreStringState] = s
		}

		if len(val.RandomString.PrefixName) > 0 {
			if prefix, ok := g.schemas.StringPrefixes[val.RandomString.PrefixName]; ok {
				s = prefix+s
			} else {
				return nil, util.NewInvalidError(fmt.Sprintf("prefix does not exist: %s",
					val.RandomString.PrefixName))
			}
		}
		if len(val.RandomString.SuffixName) > 0 {
			if suffix, ok := g.schemas.StringSuffixes[val.RandomString.SuffixName]; ok {
				s += suffix
			} else {
				return nil, util.NewInvalidError(fmt.Sprintf("suffix does not exist: %s",
					val.RandomString.SuffixName))
			}
		}
		return s, nil
	case *api.Value_VocabString:
		var s string
		if len(val.VocabString.SharedName) > 0 {
			if _, ok := eventLocalstate[val.VocabString.SharedName]; ok {
				s = eventLocalstate[val.VocabString.SharedName]
			}
		}

		if len(val.VocabString.ReadStringState) > 0 {
			if state, ok := g.globalStrings[val.VocabString.ReadStringState]; ok {
				s = state
			}
		}

		if len(s) == 0 {
			idx := rand.Int() % len(val.VocabString.Vocab)
			s = val.VocabString.Vocab[idx]
		}
		if len(val.VocabString.SharedName) > 0 {
			eventLocalstate[val.VocabString.SharedName] = s
		}
		if len(val.VocabString.StoreStringState) > 0 {
			eventGlobalState[val.VocabString.StoreStringState] = s
		}
		if len(val.VocabString.PrefixName) > 0 {
			if prefix, ok := g.schemas.StringPrefixes[val.VocabString.PrefixName]; ok {
				s = prefix+s
			} else {
				return nil, util.NewInvalidError(fmt.Sprintf("prefix does not exist: %s",
					val.VocabString.PrefixName))
			}
		}
		if len(val.VocabString.SuffixName) > 0 {
			if suffix, ok := g.schemas.StringSuffixes[val.VocabString.SuffixName]; ok {
				s += suffix
			} else {
				return nil, util.NewInvalidError(fmt.Sprintf("suffix does not exist: %s",
					val.VocabString.SuffixName))
			}
		}
		return s, nil
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
	case *api.Value_Counter:
		v := g.globalCounter[val.Counter.CounterName]
		g.globalCounter[val.Counter.CounterName]++
		return v, nil
	case *api.Value_Boolean:
		if rand.Float64() >= float64(0.5) {
			return true, nil
		} else {
			return false, nil
		}
	case *api.Value_Dict:
		out := make(map[string]interface{})
		for k, v := range val.Dict.Kvs {
			out[k], err = g.generateValue(v, eventLocalstate, eventGlobalState)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case *api.Value_List:
		var length int32
		if val.List.MaxLen == val.List.MinLen {
			length = 1
		} else {
			length = int32(rand.Float64() * float64(val.List.MaxLen - val.List.MinLen)) + val.List.MinLen
		}
		slice := make([]interface{}, length)
		for i := int32(0); i < length; i++ {
			slice[i], err = g.generateValue(val.List.Value, eventLocalstate, eventGlobalState)
			if err != nil {
				return nil, err
			}
		}
		return slice, nil
	case *api.Value_Reference:
		if _, ok := g.references[val.Reference.SchemaName]; !ok {
			return nil, util.NewInvalidError(fmt.Sprintf("schema '%s' has no valid references",
				val.Reference.SchemaName))
		}
		return g.references[val.Reference.SchemaName].Pop(val.Reference.Path)
	}
	return nil, nil
}

func (g *GeneratorState) Generate() (map[string]interface{}, error) {
	var err error
	var schema *api.Schema
	var deficientSchemas []string
	out := make(map[string]interface{})

	// Use dependency graph to determine if there is enough budget for references.  If not, generate
	// a schema that will free up budget
	for {
		if len(deficientSchemas) > 0 {
			for _, v := range g.schemas.Schemas {
				if strings.Compare(v.Name, deficientSchemas[0])  == 0 {
					schema = v
				}
			}
			deficientSchemas = deficientSchemas[1:]
			break
		} else {
			schema = g.drawSchema()
		}

		isOrphaned := g.referenceGraph.Orphaned(schema.Name)
		if isOrphaned {
			break
		}

		backEdges, err := g.referenceGraph.GetParents(schema.Name)
		if err != nil {
			return nil, err
		}

		// Ensure each parent has at least one available reference per path
		// Yes, this is a bit greedy, since the source schema may not depend
		// on all references, but it guarantees we will not get stuck
		isDeficient := false
		for _, parentVertex := range backEdges {

			// Parent has never generated values
			if _, ok := g.references[parentVertex]; !ok {
				g.references[parentVertex] = &SchemaReferences{
					make(map[string]*ReferenceValues),
				}
				isDeficient = true
			}
			for _, path := range g.references[parentVertex].Paths() {
				if g.references[parentVertex].HasPath(path) {
					if g.references[parentVertex].Num(path) > 0 {
						continue
					} else {
						// Need to generate values for schema.Name
						isDeficient = true
						break
					}
				} else {
					// Need to generate values for schema.Name
					isDeficient = true
					break
				}
			}
			if len(g.references[parentVertex].Paths()) == 0 || isDeficient {
				deficientSchemas = append(deficientSchemas, parentVertex)
			}
		}
		if !isDeficient {
			break
		}
	}

	eventLocalState := make(map[string]string)
	eventGlobalState := make(map[string]string)
	for k, v := range schema.Root.Kvs {
		out[k], err = g.generateValue(v, eventLocalState, eventGlobalState)
		if _, ok := g.references[schema.Name]; !ok {
			g.references[schema.Name] = &SchemaReferences{
				make(map[string]*ReferenceValues),
			}
		}
		switch valSpec := v.Values.(type) {
		case *api.Value_RandomString:
			g.references[schema.Name].Add(k, out[k], int(valSpec.RandomString.MaxRef))
		case *api.Value_VocabString:
			g.references[schema.Name].Add(k, out[k], int(valSpec.VocabString.MaxRef))
		case *api.Value_RandomNumeric:
			g.references[schema.Name].Add(k, out[k], int(valSpec.RandomNumeric.MaxRef))
		default:
			g.references[schema.Name].Add(k, out[k], math.MaxInt32)
		}
		if err != nil {
			return nil, err
		}
	}

	for k,v := range eventGlobalState {
		g.globalStrings[k] = v
	}

	return out, nil
}

func (g *GeneratorState) generateDependencyGraph(thisSchema string, value *api.Value) error {
	var err error
	switch val := value.Values.(type) {
	case *api.Value_Dict:
		for _, v := range val.Dict.Kvs {
			err = g.generateDependencyGraph(thisSchema, v)
			if err != nil {
				return err
			}
		}
	case *api.Value_List:
		return g.generateDependencyGraph(thisSchema, val.List.Value)
	case *api.Value_Reference:
		g.referenceGraph.AddEdge(val.Reference.SchemaName, thisSchema)
	}
	return nil
}

func (g *GeneratorState) validateDependencies() error {
	var err error

	for _, schema := range g.schemas.Schemas	{
		for _, v := range schema.Root.Kvs {
			err = g.generateDependencyGraph(schema.Name, v)
			if err != nil {
				return err
			}
		}
	}

	sortedDeps, err := TopSort(g.referenceGraph)
	if err != nil {
		return err
	}
	g.sortedDependencies = sortedDeps
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

type ThreadStats struct {
	min time.Duration
	max time.Duration
	sum time.Duration
	num int
}

func NewThreadStats() *ThreadStats {
	return &ThreadStats{
		min: math.MaxInt64,
		max: -1,
	}
}

func (t *ThreadStats) Update(elapsed time.Duration) {
	t.num++
	t.sum += elapsed
	if t.min > elapsed {
		t.min = elapsed
	}
	if t.max < elapsed {
		t.max = elapsed
	}
}

func (t *ThreadStats) MinMs() float64 {
	return float64(t.min.Milliseconds())
}

func (t *ThreadStats) MaxMs() float64 {
	return float64(t.max.Milliseconds())
}

func (t *ThreadStats) AvgMs() float64 {
	return float64(t.sum.Milliseconds()) / float64(t.num)
}

func (t *ThreadStats) Num() int {
	return t.num
}

func (t *ThreadStats) String() string {
	return fmt.Sprintf("min: %f ms, max: %f ms, avg: %f ms", t.MinMs(), t.MaxMs(), t.AvgMs())
}

type Stats struct {
	threadStats []*ThreadStats
	startTime time.Time
	endTime time.Time
	num int
}

func NewStats(numThreads int) *Stats {
	threadStats := make([]*ThreadStats, numThreads)
	for i := 0; i < numThreads; i++ {
		threadStats[i] = NewThreadStats()
	}
	return &Stats{
		threadStats: threadStats,
	}
}

func (s *Stats) AddLatency(threadNum int, elapsed time.Duration) {
	s.num++
	s.threadStats[threadNum].Update(elapsed)
}

func (s *Stats) Start() {
	s.startTime = time.Now()
}

func (s *Stats) Finish() string {
	str := ""
	s.endTime = time.Now()

	str += fmt.Sprintf("Total duration: %d ms\n", s.endTime.Sub(s.startTime).Milliseconds())
	str += fmt.Sprintf("RPS: %f\n", float64(s.num) / float64(s.endTime.Sub(s.startTime).Seconds()))
	for i, threadStat := range s.threadStats {
		str += fmt.Sprintf("Thread %d: %s\n", i, threadStat.String())
	}
	return str
}

type CommandState struct {
	schemas *api.Schemas
	sender Sender
	numEvents int32
	numThreads int
	silent bool
	stats *Stats
}

func usage(msg string, exitCode int) {
	fmt.Println(msg)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

type Sender interface {
	Send([]byte) error
}

type FileSender struct {
	file *os.File
	lock *sync.Mutex
}

func NewFileSender(filename string, lock *sync.Mutex) (*FileSender, error) {
	var err error
	var fp *os.File
	if len(filename) == 0 {
		fp = os.Stdout
	} else {
		fp, err = os.OpenFile(filename, os.O_CREATE | os.O_RDWR, os.ModePerm | 0644)
		if err != nil {
			return nil, err
		}
	}
	return &FileSender{
		fp,
		lock,

	}, nil
}

func (s *FileSender) Send(rawJson []byte) error {
	pretty := bytes.NewBuffer([]byte{})
	err := json.Indent(pretty, rawJson, "", "\t")
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err = fmt.Fprintln(s.file, pretty.String())
	if err != nil {
		return err
	}
	return nil
}

func isUrl(url string) bool {
	reStr := "^(https?:\\/\\/)?[0-9a-zA-Z.]+"
	re, _ := regexp.Compile(reStr)
	return re.Match([]byte(url))
}

func NewHttpSender(url string) (*HttpSender, error) {
	if !isUrl(url) {
		return nil, util.NewInvalidError(fmt.Sprintf("'%s' is not a valid URL", url))
	}

	return &HttpSender{
		url,
	}, nil
}

type HttpSender struct {
	url string
}

func (s *HttpSender) Send(rawJson []byte) error {
	resp, err := http.DefaultClient.Post(s.url, "application/json", bytes.NewBuffer(rawJson))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func parseArgs() *CommandState {
	var err error
	args := &CommandState{}
	schemaPtr := flag.String("schema", "", "path to config file schemas")
	numEventsPtr := flag.Int("numEvents", 100, "number of events to generate")
	numThreadsPtr := flag.Int("numThreads", 1, "number of threads used to generate events")
	silentPtr := flag.Bool("silent", false, "if specified, will not print statistics")
	outputUsage := `destination to send the events:
- File: a '/' delimited path ("" will go to stdout)
- Endpoint: a http://host:port/path endpoint
`
	outputPtr := flag.String("output", "", outputUsage)
	flag.Parse()
	args.numEvents = int32(*numEventsPtr)
	schemaDef, err := os.Open(*schemaPtr)
	if err != nil {
		usage(err.Error(), 1)
	}

	schemaDefBytes, err := ioutil.ReadAll(schemaDef)
	if err != nil {
		usage(err.Error(), 1)
	}

	args.schemas, err = SchemasFromJson(schemaDefBytes)
	if err != nil {
		usage(err.Error(), 1)
	}

	args.numThreads = *numThreadsPtr

	args.stats = NewStats(args.numThreads)

	args.silent = *silentPtr

	if isUrl(*outputPtr) {
		args.sender, err = NewHttpSender(*outputPtr)
		if err != nil {
			usage(err.Error(), 1)
		}
	} else {
		args.sender, err = NewFileSender(*outputPtr, &sync.Mutex{})
		if err != nil {
			usage(err.Error(), 1)
		}
	}

	return args
}

func generator(threadNum int, args *CommandState) error {
	generatorState, err := NewGeneratorState(args.schemas)
	if err != nil {
		return err
	}

	atomic.AddInt32(&args.numEvents, -1)

	for args.numEvents >= 0 {
		out, err := generatorState.Generate()
		if err != nil {
			return err
		}

		rawJson, err := util.MapToJson(out)
		if err != nil {
			return err
		}

		start := time.Now()
		err = args.sender.Send(rawJson)
		if err != nil {
			return err
		}
		elapsed := time.Now().Sub(start)
		args.stats.AddLatency(threadNum, elapsed)
		atomic.AddInt32(&args.numEvents, -1)
	}
	return nil
}

func main() {
	args := parseArgs()
	wg := sync.WaitGroup{}
	wg.Add(args.numThreads)

	args.stats.Start()
	for i := 0; i < args.numThreads; i++ {
		go func(threadNum int) {
			_ = generator(threadNum, args)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if !args.silent {
		fmt.Println(args.stats.Finish())
	}
}
