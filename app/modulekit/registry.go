package modulekit

import (
	"fmt"
	"sort"
	"strings"

	apic "github.com/imfact-labs/currency-model/api"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/ps"
)

type ModelModule interface {
	ID() string
	Register(*Registry) error
}

type OperationProcessors struct {
	Name      ps.Name
	Func      ps.Func
	SupportsA bool
	SupportsB bool
}

type APIRoute struct {
	Path    string
	Methods []string
}

type CLICommand struct {
	Key         string
	Description string
}

type APIHandlerInitializer struct {
	Key      string
	Register func(*apic.Handlers, bool)
}

type ModuleEntry struct {
	ID                  string
	Hinters             []encoder.DecodeDetail
	SupportedFacts      []encoder.DecodeDetail
	OperationProcessors []OperationProcessors
	APIRoutes           []APIRoute
	APIHandlers         []APIHandlerInitializer
	CLICommands         []CLICommand
}

type Registry struct {
	modules map[string]*ModuleEntry
	order   []string

	hinterOwners    map[string]string
	factOwners      map[string]string
	processorOwners map[ps.Name]string
	routeOwners     map[string]string
	apiOwners       map[string]string
	cliOwners       map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		modules:         map[string]*ModuleEntry{},
		order:           []string{},
		hinterOwners:    map[string]string{},
		factOwners:      map[string]string{},
		processorOwners: map[ps.Name]string{},
		routeOwners:     map[string]string{},
		apiOwners:       map[string]string{},
		cliOwners:       map[string]string{},
	}
}

func (r *Registry) Register(module ModelModule) error {
	if module == nil {
		return fmt.Errorf("nil module")
	}

	id := strings.TrimSpace(module.ID())
	if id == "" {
		return fmt.Errorf("empty module id")
	}

	if _, exists := r.modules[id]; exists {
		return fmt.Errorf("duplicated module id, %q", id)
	}

	r.modules[id] = &ModuleEntry{ID: id}
	r.order = append(r.order, id)

	return module.Register(r)
}

func (r *Registry) ModuleIDs() []string {
	ids := make([]string, 0, len(r.modules))
	for id := range r.modules {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}

func (r *Registry) Module(id string) (ModuleEntry, bool) {
	entry, found := r.modules[id]
	if !found {
		return ModuleEntry{}, false
	}

	return ModuleEntry{
		ID:                  entry.ID,
		Hinters:             append([]encoder.DecodeDetail(nil), entry.Hinters...),
		SupportedFacts:      append([]encoder.DecodeDetail(nil), entry.SupportedFacts...),
		OperationProcessors: append([]OperationProcessors(nil), entry.OperationProcessors...),
		APIRoutes:           append([]APIRoute(nil), entry.APIRoutes...),
		APIHandlers:         append([]APIHandlerInitializer(nil), entry.APIHandlers...),
		CLICommands:         append([]CLICommand(nil), entry.CLICommands...),
	}, true
}

func (r *Registry) Entries() []ModuleEntry {
	entries := make([]ModuleEntry, 0, len(r.order))

	for i := range r.order {
		if entry, found := r.Module(r.order[i]); found {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (r *Registry) ValidateModuleContract(id string) error {
	entry, found := r.modules[id]
	if !found {
		return fmt.Errorf("module not found, %q", id)
	}

	if len(entry.Hinters) < 1 {
		return fmt.Errorf("module %q: empty hinters registration", id)
	}
	if len(entry.SupportedFacts) < 1 {
		return fmt.Errorf("module %q: empty supported facts registration", id)
	}
	if len(entry.OperationProcessors) < 1 {
		return fmt.Errorf("module %q: empty operation processors registration", id)
	}
	if len(entry.APIRoutes) < 1 {
		return fmt.Errorf("module %q: empty api routes registration", id)
	}
	if len(entry.APIHandlers) < 1 {
		return fmt.Errorf("module %q: empty api handlers registration", id)
	}
	if len(entry.CLICommands) < 1 {
		return fmt.Errorf("module %q: empty cli commands registration", id)
	}

	return nil
}

func (r *Registry) AddHinters(moduleID string, details ...encoder.DecodeDetail) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range details {
		key := details[i].Hint.String()
		if owner, found := r.hinterOwners[key]; found {
			return fmt.Errorf("duplicated hinter %q; owner=%q, conflict=%q", key, owner, moduleID)
		}
		r.hinterOwners[key] = moduleID
		entry.Hinters = append(entry.Hinters, details[i])
	}

	return nil
}

func (r *Registry) AddSupportedFacts(moduleID string, details ...encoder.DecodeDetail) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range details {
		key := details[i].Hint.String()
		if owner, found := r.factOwners[key]; found {
			return fmt.Errorf("duplicated supported fact %q; owner=%q, conflict=%q", key, owner, moduleID)
		}
		r.factOwners[key] = moduleID
		entry.SupportedFacts = append(entry.SupportedFacts, details[i])
	}

	return nil
}

func (r *Registry) AddOperationProcessors(moduleID string, processors ...OperationProcessors) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range processors {
		if err := validateOperationProcessors(processors[i]); err != nil {
			return fmt.Errorf("module %q: %w", moduleID, err)
		}
		if owner, found := r.processorOwners[processors[i].Name]; found {
			return fmt.Errorf(
				"duplicated operation processor %q; owner=%q, conflict=%q",
				processors[i].Name,
				owner,
				moduleID,
			)
		}

		r.processorOwners[processors[i].Name] = moduleID
		entry.OperationProcessors = append(entry.OperationProcessors, processors[i])
	}

	return nil
}

func (r *Registry) AddAPIRoutes(moduleID string, routes ...APIRoute) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range routes {
		key, err := routeKey(routes[i])
		if err != nil {
			return fmt.Errorf("module %q: %w", moduleID, err)
		}
		if owner, found := r.routeOwners[key]; found {
			return fmt.Errorf("duplicated api route %q; owner=%q, conflict=%q", key, owner, moduleID)
		}

		r.routeOwners[key] = moduleID
		entry.APIRoutes = append(entry.APIRoutes, routes[i])
	}

	return nil
}

func (r *Registry) AddAPIHandlers(moduleID string, handlers ...APIHandlerInitializer) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range handlers {
		key := strings.TrimSpace(handlers[i].Key)
		if key == "" {
			return fmt.Errorf("module %q: empty api handler key", moduleID)
		}
		if handlers[i].Register == nil {
			return fmt.Errorf("module %q: nil api handler register func for %q", moduleID, key)
		}

		if owner, found := r.apiOwners[key]; found {
			return fmt.Errorf("duplicated api handler %q; owner=%q, conflict=%q", key, owner, moduleID)
		}

		handlers[i].Key = key
		r.apiOwners[key] = moduleID
		entry.APIHandlers = append(entry.APIHandlers, handlers[i])
	}

	return nil
}

func (r *Registry) AddCLICommands(moduleID string, commands ...CLICommand) error {
	entry, err := r.requireModule(moduleID)
	if err != nil {
		return err
	}

	for i := range commands {
		key := strings.TrimSpace(commands[i].Key)
		if key == "" {
			return fmt.Errorf("module %q: empty cli command key", moduleID)
		}
		if owner, found := r.cliOwners[key]; found {
			return fmt.Errorf("duplicated cli command %q; owner=%q, conflict=%q", key, owner, moduleID)
		}

		r.cliOwners[key] = moduleID
		entry.CLICommands = append(entry.CLICommands, commands[i])
	}

	return nil
}

func (r *Registry) requireModule(id string) (*ModuleEntry, error) {
	moduleID := strings.TrimSpace(id)
	if moduleID == "" {
		return nil, fmt.Errorf("empty module id")
	}

	entry, found := r.modules[moduleID]
	if !found {
		return nil, fmt.Errorf("module not registered, %q", moduleID)
	}

	return entry, nil
}

func validateOperationProcessors(p OperationProcessors) error {
	if strings.TrimSpace(p.Name.String()) == "" {
		return fmt.Errorf("empty operation processor name")
	}
	if p.Func == nil {
		return fmt.Errorf("nil operation processor func for %q", p.Name)
	}
	if !p.SupportsA && !p.SupportsB {
		return fmt.Errorf("operation processor %q must support A or B", p.Name)
	}

	return nil
}

func routeKey(route APIRoute) (string, error) {
	path := strings.TrimSpace(route.Path)
	if path == "" {
		return "", fmt.Errorf("empty api route path")
	}

	if len(route.Methods) < 1 {
		return path + "|*", nil
	}

	methods := make([]string, len(route.Methods))
	for i := range route.Methods {
		m := strings.ToUpper(strings.TrimSpace(route.Methods[i]))
		if m == "" {
			return "", fmt.Errorf("empty api route method for %q", path)
		}
		methods[i] = m
	}

	sort.Strings(methods)

	return path + "|" + strings.Join(methods, ","), nil
}
