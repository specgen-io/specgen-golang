package service

import (
	"github.com/specgen-io/specgen-golang/v2/empty"
	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/openapi"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
)

func GenerateService(specification *spec.Spec, jsonmode, server, moduleName, swaggerPath, generatePath, servicesPath string) *generator.Sources {
	sources := generator.NewSources()

	modules := NewModules(moduleName, generatePath, servicesPath, specification)
	generator := NewGenerator(jsonmode, server, modules)

	sources.AddGenerated(empty.GenerateEmpty(generator.Modules.Empty))
	sources.AddGenerated(generator.EnumsHelperFunctions())
	sources.AddGenerated(generator.ResponseHelperFunctions())
	sources.AddGenerated(generator.CheckContentType())
	sources.AddGenerated(generator.GenerateParamsParser())
	sources.AddGenerated(generator.GenerateFormDataParamsParser())
	sources.AddGenerated(generator.GenerateFormUrlencodedParamsParser())

	sources.AddGeneratedAll(generator.ErrorModels(specification.HttpErrors))
	sources.AddGeneratedAll(generator.HttpErrors(&specification.HttpErrors.Responses))

	sources.AddGenerated(generator.RootRouting(specification))
	sources.AddGenerated(generator.GenerateUrlParamsCtor())
	for _, version := range specification.Versions {
		sources.AddGeneratedAll(generator.Routings(&version))
		sources.AddGeneratedAll(generator.ServicesInterfaces(&version))
		sources.AddGeneratedAll(generator.Models(&version))
	}

	if swaggerPath != "" {
		sources.AddGenerated(openapi.GenerateOpenapi(specification, swaggerPath))
	}

	if servicesPath != "" {
		for _, version := range specification.Versions {
			sources.AddScaffoldedAll(generator.ServicesImpls(&version))
		}
	}

	return sources
}
