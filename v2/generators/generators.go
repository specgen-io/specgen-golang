package generators

import (
	"github.com/specgen-io/specgen-golang/v2/client"
	"github.com/specgen-io/specgen-golang/v2/goven/github.com/specgen-io/specgen/generator/v2"
	"github.com/specgen-io/specgen-golang/v2/goven/github.com/specgen-io/specgen/spec/v2"
	"github.com/specgen-io/specgen-golang/v2/models"
	"github.com/specgen-io/specgen-golang/v2/service"
)

var Models = generator.Generator{
	"models-go",
	"Go Models",
	"Generate Go models source code",
	[]generator.GeneratorArg{
		{Arg: generator.ArgSpecFile, Required: true},
		{Arg: generator.ArgModuleName, Required: true},
		{Arg: generator.ArgGeneratePath, Required: true},
	},
	func(specification *spec.Spec, params generator.GeneratorArgsValues) *generator.Sources {
		return models.GenerateModels(specification, params[generator.ArgModuleName], params[generator.ArgGeneratePath])
	},
}

var Client = generator.Generator{
	"client-go",
	"Go Client",
	"Generate Go client source code",
	[]generator.GeneratorArg{
		{Arg: generator.ArgSpecFile, Required: true},
		{Arg: generator.ArgModuleName, Required: true},
		{Arg: generator.ArgGeneratePath, Required: true},
	},
	func(specification *spec.Spec, params generator.GeneratorArgsValues) *generator.Sources {
		return client.GenerateClient(specification, params[generator.ArgModuleName], params[generator.ArgGeneratePath])
	},
}

var Service = generator.Generator{
	"service-go",
	"Go Service",
	"Generate Go service source code",
	[]generator.GeneratorArg{
		{Arg: generator.ArgSpecFile, Required: true},
		{Arg: generator.ArgModuleName, Required: true},
		{Arg: generator.ArgSwaggerPath, Required: false},
		{Arg: generator.ArgGeneratePath, Required: true},
		{Arg: generator.ArgServicesPath, Required: false},
	},
	func(specification *spec.Spec, params generator.GeneratorArgsValues) *generator.Sources {
		return service.GenerateService(specification, params[generator.ArgModuleName], params[generator.ArgSwaggerPath], params[generator.ArgGeneratePath], params[generator.ArgServicesPath])
	},
}

var All = []generator.Generator{
	Models,
	Client,
	Service,
}
