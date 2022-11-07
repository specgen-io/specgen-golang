package imports

import (
	"fmt"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
	"github.com/specgen-io/specgen-golang/v2/module"
	"github.com/specgen-io/specgen-golang/v2/types"
	"github.com/specgen-io/specgen-golang/v2/writer"
	"sort"
)

type imports struct {
	imports map[string]string
}

func New() *imports {
	return &imports{imports: make(map[string]string)}
}

func (self *imports) Module(module module.Module) *imports {
	self.Add(module.Package)
	return self
}

func (self *imports) ModuleAliased(module module.Module) *imports {
	if module.Alias == "" {
		panic(fmt.Sprintf(`module %s does not have alias and can't imported as aliased'`, module.Package))
	}
	self.AddAliased(module.Package, module.Alias)
	return self
}

func (self *imports) Add(theImport string) *imports {
	self.imports[theImport] = ""
	return self
}

func (self *imports) AddAliased(theImport string, alias string) *imports {
	self.imports[theImport] = alias
	return self
}

func (self *imports) Write(w *writer.Writer) {
	if len(self.imports) > 0 {
		imports := make([]string, 0, len(self.imports))
		for theImport := range self.imports {
			imports = append(imports, theImport)
		}
		sort.Strings(imports)

		w.Line(`import (`)
		for _, theImport := range imports {
			alias := self.imports[theImport]
			if alias != "" {
				w.Line(`  %s "%s"`, alias, theImport)
			} else {
				w.Line(`  "%s"`, theImport)
			}
		}
		w.Line(`)`)
	}
}

func (self *imports) AddApiTypes(api *spec.Api) *imports {
	if types.ApiHasType(api, spec.TypeDate) {
		self.Add("cloud.google.com/go/civil")
	}
	if types.ApiHasType(api, spec.TypeJson) {
		self.Add("encoding/json")
	}
	if types.ApiHasType(api, spec.TypeUuid) {
		self.Add("github.com/google/uuid")
	}
	if types.ApiHasType(api, spec.TypeDecimal) {
		self.Add("github.com/shopspring/decimal")
	}
	return self
}

func (self *imports) AddModelsTypes(models []*spec.NamedModel) *imports {
	self.Add("errors")
	self.Add("encoding/json")
	if types.VersionModelsHasType(models, spec.TypeDate) {
		self.Add("cloud.google.com/go/civil")
	}
	if types.VersionModelsHasType(models, spec.TypeUuid) {
		self.Add("github.com/google/uuid")
	}
	if types.VersionModelsHasType(models, spec.TypeDecimal) {
		self.Add("github.com/shopspring/decimal")
	}
	return self
}
