package service

import (
	"github.com/specgen-io/specgen-go/v2/generator"
	"github.com/specgen-io/specgen-go/v2/openapi"
	"github.com/specgen-io/specgen-go/v2/spec"
	"github.com/specgen-io/specgen-go/v2/gen/java/packages"
)

func Generate(specification *spec.Spec, jsonlib, server, packageName, swaggerPath, generatePath, servicesPath string) *generator.Sources {
	sources := generator.NewSources()

	if packageName == "" {
		packageName = specification.Name.SnakeCase()
	}

	mainPackage := packages.New(generatePath, packageName)

	generator := NewGenerator(jsonlib, server)

	jsonPackage := mainPackage.Subpackage("json")
	for _, version := range specification.Versions {
		versionPackage := mainPackage.Subpackage(version.Version.FlatCase())

		modelsVersionPackage := versionPackage.Subpackage("models")
		sources.AddGeneratedAll(generator.Models.VersionModels(&version, modelsVersionPackage, jsonPackage))

		serviceVersionPackage := versionPackage.Subpackage("services")
		sources.AddGeneratedAll(generator.ServicesInterfaces(&version, serviceVersionPackage, modelsVersionPackage))

		controllerVersionPackage := versionPackage.Subpackage("controllers")
		sources.AddGeneratedAll(generator.Server.ServicesControllers(&version, mainPackage, controllerVersionPackage, jsonPackage, modelsVersionPackage, serviceVersionPackage))
	}

	sources.AddGeneratedAll(generator.Models.SetupLibrary(jsonPackage))

	if swaggerPath != "" {
		sources.AddGenerated(openapi.GenerateOpenapi(specification, swaggerPath))
	}

	if servicesPath != "" {
		servicesImplPackage := packages.New(servicesPath, packageName)
		for _, version := range specification.Versions {
			servicesImplVersionPath := servicesImplPackage.Subpackage("services")
			serviceImplVersionPackage := servicesImplVersionPath.Subpackage(version.Version.FlatCase())

			versionPackage := mainPackage.Subpackage(version.Version.FlatCase())
			modelsVersionPackage := versionPackage.Subpackage("models")
			serviceVersionPackage := versionPackage.Subpackage("services")

			sources.AddScaffoldedAll(generator.ServicesImplementations(&version, serviceImplVersionPackage, modelsVersionPackage, serviceVersionPackage))
		}
	}

	return sources
}
