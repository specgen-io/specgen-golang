package service

import (
	"github.com/specgen-io/specgen-go/v2/spec"
	"github.com/specgen-io/specgen-go/v2/gen/golang/models"
	"github.com/specgen-io/specgen-go/v2/gen/golang/module"
	"github.com/specgen-io/specgen-go/v2/gen/golang/types"
	"github.com/specgen-io/specgen-go/v2/gen/openapi"
	"github.com/specgen-io/specgen-go/v2/generator"
)

func GenerateService(specification *spec.Spec, moduleName string, swaggerPath string, generatePath string, servicesPath string) *generator.Sources {
	sources := generator.NewSources()

	rootModule := module.New(moduleName, generatePath)

	emptyModule := rootModule.Submodule("empty")
	sources.AddGenerated(types.GenerateEmpty(emptyModule))
	sources.AddGenerated(generateSpecRouting(specification, rootModule))

	for _, version := range specification.Versions {
		versionModule := rootModule.Submodule(version.Version.FlatCase())
		modelsModule := versionModule.Submodule(types.ModelsPackage)

		sources.AddGenerated(generateParamsParser(versionModule, modelsModule))
		sources.AddGeneratedAll(generateRoutings(&version, versionModule, modelsModule))
		sources.AddGeneratedAll(generateServiceInterfaces(&version, versionModule, modelsModule, emptyModule))
		sources.AddGeneratedAll(models.GenerateVersionModels(&version, modelsModule))
	}

	if swaggerPath != "" {
		sources.AddGenerated(openapi.GenerateOpenapi(specification, swaggerPath))
	}

	if servicesPath != "" {
		rootServicesModule := module.New(moduleName, servicesPath)
		for _, version := range specification.Versions {
			versionServicesModule := rootServicesModule.Submodule(version.Version.FlatCase())
			versionModule := rootModule.Submodule(version.Version.FlatCase())
			modelsModule := versionModule.Submodule(types.ModelsPackage)
			sources.AddScaffoldedAll(generateServiceImplementations(&version, versionModule, modelsModule, versionServicesModule))
		}
	}

	return sources
}
