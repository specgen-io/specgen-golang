package client

import (
	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
	"github.com/specgen-io/specgen-golang/v2/models"
	"github.com/specgen-io/specgen-golang/v2/module"
	"github.com/specgen-io/specgen-golang/v2/types"
)

func GenerateClient(specification *spec.Spec, moduleName string, generatePath string) *generator.Sources {
	sources := generator.NewSources()

	modules := models.NewModules(moduleName, generatePath, specification)
	generator := NewGenerator(modules)

	rootModule := module.New(moduleName, generatePath)

	sources.AddGenerated(generator.Models.GenerateEnumsHelperFunctions())

	emptyModule := rootModule.Submodule("empty")
	sources.AddGenerated(types.GenerateEmpty(emptyModule))

	convertModule := rootModule.Submodule("convert")
	sources.AddGenerated(generateConverter(convertModule))

	responseModule := rootModule.Submodule("response")
	sources.AddGenerated(generateResponseFunctions(responseModule))

	errorsModule := rootModule.Submodule("httperrors")
	errorsModelsModule := errorsModule.Submodule(types.ErrorsModelsPackage)
	sources.AddGenerated(generator.Models.GenerateErrorModels(specification.HttpErrors))
	sources.AddGenerated(generator.HttpErrors(errorsModule, errorsModelsModule, &specification.HttpErrors.Responses))

	for _, version := range specification.Versions {
		versionModule := rootModule.Submodule(version.Name.FlatCase())
		modelsModule := versionModule.Submodule(types.VersionModelsPackage)
		sources.AddGenerated(generator.Models.GenerateVersionModels(&version))
		sources.AddGeneratedAll(generator.Client.GenerateClientsImplementations(&version, versionModule, convertModule, emptyModule, errorsModule, errorsModelsModule, modelsModule, responseModule))
	}
	return sources
}
