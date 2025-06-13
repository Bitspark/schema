package export

import (
	"fmt"
	"sort"
	"sync"

	"defs.dev/schema/core"
	"defs.dev/schema/visit/export/base"
)

// RegistryImpl implements the GeneratorRegistry interface.
type RegistryImpl struct {
	mu         sync.RWMutex
	generators map[string]Generator
	factories  map[string]GeneratorFactoryFunc
}

// NewGeneratorRegistry creates a new GeneratorRegistry.
func NewGeneratorRegistry() GeneratorRegistry {
	return &RegistryImpl{
		generators: make(map[string]Generator),
		factories:  make(map[string]GeneratorFactoryFunc),
	}
}

// Register adds a generator with a given name.
func (r *RegistryImpl) Register(name string, generator Generator) error {
	if name == "" {
		return fmt.Errorf("generator name cannot be empty")
	}
	if generator == nil {
		return fmt.Errorf("generator cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.generators[name] = generator
	return nil
}

// RegisterFactory adds a generator factory with a given name.
func (r *RegistryImpl) RegisterFactory(name string, factory GeneratorFactory) error {
	if name == "" {
		return fmt.Errorf("factory name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert GeneratorFactory to GeneratorFactoryFunc
	factoryFunc := func(options ...Option) (Generator, error) {
		// Convert options to the expected format
		var anyOptions []any
		for _, opt := range options {
			anyOptions = append(anyOptions, opt)
		}
		return factory(anyOptions...)
	}

	r.factories[name] = factoryFunc
	return nil
}

// Get retrieves a generator by name.
func (r *RegistryImpl) Get(name string) (Generator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First check for registered generators
	if generator, exists := r.generators[name]; exists {
		return generator, true
	}

	// Then check for factories and create a default generator
	if factory, exists := r.factories[name]; exists {
		generator, err := factory() // Create with no options
		if err != nil {
			return nil, false
		}
		return generator, true
	}

	return nil, false
}

// List returns all registered generator names.
func (r *RegistryImpl) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.generators)+len(r.factories))

	// Add generator names
	for name := range r.generators {
		names = append(names, name)
	}

	// Add factory names (if not already a generator)
	for name := range r.factories {
		if _, exists := r.generators[name]; !exists {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names
}

// Generate uses a specific generator to produce output.
func (r *RegistryImpl) Generate(generatorName string, schema core.Schema) ([]byte, error) {
	generator, exists := r.Get(generatorName)
	if !exists {
		return nil, fmt.Errorf("generator not found: %s", generatorName)
	}

	return generator.Generate(schema)
}

// GenerateAll produces output using all registered generators.
func (r *RegistryImpl) GenerateAll(schema core.Schema) (map[string][]byte, error) {
	names := r.List()
	results := make(map[string][]byte)
	errors := base.NewErrorCollector()

	for _, name := range names {
		output, err := r.Generate(name, schema)
		if err != nil {
			errors.Addf("generator %s failed: %v", name, err)
			continue
		}
		results[name] = output
	}

	if errors.HasErrors() {
		return results, errors.Error()
	}

	return results, nil
}

// Remove unregisters a generator.
func (r *RegistryImpl) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := false

	if _, exists := r.generators[name]; exists {
		delete(r.generators, name)
		removed = true
	}

	if _, exists := r.factories[name]; exists {
		delete(r.factories, name)
		removed = true
	}

	return removed
}

// GetWithFactory retrieves a generator by name, creating it with the given options if it's a factory.
func (r *RegistryImpl) GetWithFactory(name string, options ...Option) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First check for registered generators
	if generator, exists := r.generators[name]; exists {
		return generator, nil
	}

	// Then check for factories
	if factory, exists := r.factories[name]; exists {
		return factory(options...)
	}

	return nil, fmt.Errorf("generator or factory not found: %s", name)
}

// HasGenerator returns true if a generator or factory is registered for the given name.
func (r *RegistryImpl) HasGenerator(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, hasGenerator := r.generators[name]
	_, hasFactory := r.factories[name]

	return hasGenerator || hasFactory
}

// Clone creates a copy of the registry.
func (r *RegistryImpl) Clone() GeneratorRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := &RegistryImpl{
		generators: make(map[string]Generator),
		factories:  make(map[string]GeneratorFactoryFunc),
	}

	// Copy generators
	for name, generator := range r.generators {
		clone.generators[name] = generator
	}

	// Copy factories
	for name, factory := range r.factories {
		clone.factories[name] = factory
	}

	return clone
}

// Stats returns statistics about the registry.
func (r *RegistryImpl) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RegistryStats{
		TotalGenerators: len(r.generators),
		TotalFactories:  len(r.factories),
		GeneratorNames:  r.getGeneratorNames(),
		FactoryNames:    r.getFactoryNames(),
	}
}

// getGeneratorNames returns generator names (must be called with lock held).
func (r *RegistryImpl) getGeneratorNames() []string {
	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getFactoryNames returns factory names (must be called with lock held).
func (r *RegistryImpl) getFactoryNames() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegistryStats contains statistics about a generator registry.
type RegistryStats struct {
	TotalGenerators int
	TotalFactories  int
	GeneratorNames  []string
	FactoryNames    []string
}

// BatchGenerator provides advanced batch generation capabilities.
type BatchGenerator struct {
	registry GeneratorRegistry
	options  BatchOptions
}

// BatchOptions configures batch generation behavior.
type BatchOptions struct {
	// ContinueOnError determines whether to continue generation if one generator fails
	ContinueOnError bool

	// Parallel enables parallel generation
	Parallel bool

	// MaxConcurrency limits the number of concurrent generators (0 = unlimited)
	MaxConcurrency int

	// IncludeStats includes generation statistics in results
	IncludeStats bool

	// Generators specifies which generators to use (empty = all)
	Generators []string
}

// NewBatchGenerator creates a new BatchGenerator.
func NewBatchGenerator(registry GeneratorRegistry, options BatchOptions) *BatchGenerator {
	return &BatchGenerator{
		registry: registry,
		options:  options,
	}
}

// Generate performs batch generation across multiple generators.
func (b *BatchGenerator) Generate(schema core.Schema) *BatchGenerationResult {
	generators := b.options.Generators
	if len(generators) == 0 {
		generators = b.registry.List()
	}

	result := &BatchGenerationResult{
		Results: make(map[string]*GenerationResult),
		Errors:  make(map[string]error),
		Summary: &GenerationSummary{
			TotalGenerators: len(generators),
		},
	}

	if b.options.Parallel && len(generators) > 1 {
		b.generateParallel(schema, generators, result)
	} else {
		b.generateSequential(schema, generators, result)
	}

	// Update summary
	result.Summary.SuccessfulGenerators = len(result.Results)
	result.Summary.FailedGenerators = len(result.Errors)

	// Count total warnings
	totalWarnings := 0
	for _, genResult := range result.Results {
		totalWarnings += len(genResult.Warnings)
	}
	result.Summary.TotalWarnings = totalWarnings

	return result
}

// generateSequential performs sequential generation.
func (b *BatchGenerator) generateSequential(schema core.Schema, generators []string, result *BatchGenerationResult) {
	for _, name := range generators {
		output, err := b.registry.Generate(name, schema)
		if err != nil {
			result.Errors[name] = err
			if !b.options.ContinueOnError {
				break
			}
			continue
		}

		generator, _ := b.registry.Get(name)
		result.Results[name] = &GenerationResult{
			Output:   output,
			Format:   generator.Format(),
			Metadata: make(map[string]any),
		}
	}
}

// generateParallel performs parallel generation.
func (b *BatchGenerator) generateParallel(schema core.Schema, generators []string, result *BatchGenerationResult) {
	concurrency := b.options.MaxConcurrency
	if concurrency <= 0 || concurrency > len(generators) {
		concurrency = len(generators)
	}

	// Create channels for work distribution
	jobs := make(chan string, len(generators))
	results := make(chan generationJob, len(generators))

	// Start workers
	for i := 0; i < concurrency; i++ {
		go b.generateWorker(schema, jobs, results)
	}

	// Send jobs
	for _, name := range generators {
		jobs <- name
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(generators); i++ {
		job := <-results
		if job.err != nil {
			result.Errors[job.name] = job.err
		} else {
			result.Results[job.name] = job.result
		}
	}
}

// generationJob represents a single generation job result.
type generationJob struct {
	name   string
	result *GenerationResult
	err    error
}

// generateWorker is a worker goroutine for parallel generation.
func (b *BatchGenerator) generateWorker(schema core.Schema, jobs <-chan string, results chan<- generationJob) {
	for name := range jobs {
		output, err := b.registry.Generate(name, schema)
		if err != nil {
			results <- generationJob{name: name, err: err}
			continue
		}

		generator, _ := b.registry.Get(name)
		results <- generationJob{
			name: name,
			result: &GenerationResult{
				Output:   output,
				Format:   generator.Format(),
				Metadata: make(map[string]any),
			},
		}
	}
}

// DefaultRegistry is the global generator registry.
var DefaultRegistry = NewGeneratorRegistry()

// Convenience functions for the default registry

// Register registers a generator in the default registry.
func Register(name string, generator Generator) error {
	return DefaultRegistry.Register(name, generator)
}

// RegisterFactory registers a generator factory in the default registry.
func RegisterFactory(name string, factory GeneratorFactory) error {
	return DefaultRegistry.RegisterFactory(name, factory)
}

// Get retrieves a generator from the default registry.
func Get(name string) (Generator, bool) {
	return DefaultRegistry.Get(name)
}

// Generate uses the default registry to generate output.
func Generate(generatorName string, schema core.Schema) ([]byte, error) {
	return DefaultRegistry.Generate(generatorName, schema)
}

// GenerateAll uses the default registry to generate output with all generators.
func GenerateAll(schema core.Schema) (map[string][]byte, error) {
	return DefaultRegistry.GenerateAll(schema)
}

// Remove removes a generator from the default registry.
func Remove(name string) bool {
	return DefaultRegistry.Remove(name)
}

// List returns all generator names from the default registry.
func List() []string {
	return DefaultRegistry.List()
}
