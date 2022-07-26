package service

import (
	"fmt"
	"github.com/pinzolo/casee"
	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
	"github.com/specgen-io/specgen-golang/v2/models"
	"github.com/specgen-io/specgen-golang/v2/types"
	"github.com/specgen-io/specgen-golang/v2/walkers"
	"github.com/specgen-io/specgen-golang/v2/writer"
	"strings"
)

type VestigoGenerator struct {
	Types   *types.Types
	Models  models.Generator
	Modules *Modules
}

func NewVestigoGenerator(types *types.Types, models models.Generator, modules *Modules) *VestigoGenerator {
	return &VestigoGenerator{types, models, modules}
}

func (g *VestigoGenerator) Routings(version *spec.Version) []generator.CodeFile {
	files := []generator.CodeFile{}
	for _, api := range version.Http.Apis {
		files = append(files, *g.routing(&api))
	}
	return files
}

func (g *VestigoGenerator) signatureAddRouting(api *spec.Api) string {
	fullServiceInterfaceName := fmt.Sprintf("%s.%s", api.Name.SnakeCase(), serviceInterfaceName)
	return fmt.Sprintf(`%s(router *vestigo.Router, %s %s)`, g.addRoutesMethodName(api), serviceInterfaceTypeVar(api), fullServiceInterfaceName)
}

func (g *VestigoGenerator) routing(api *spec.Api) *generator.CodeFile {
	w := writer.New(g.Modules.Routing(api.InHttp.InVersion), fmt.Sprintf("%s.go", api.Name.SnakeCase()))

	w.Imports.Add("github.com/husobee/vestigo")
	w.Imports.AddAliased("github.com/sirupsen/logrus", "log")
	w.Imports.Add("net/http")
	w.Imports.Add("fmt")
	if walkers.ApiHasBodyOfKind(api, spec.BodyString) {
		w.Imports.Add("io/ioutil")
	}
	if walkers.ApiHasBodyOfKind(api, spec.BodyJson) {
		w.Imports.Add("encoding/json")
	}
	if walkers.ApiHasBodyOfKind(api, spec.BodyJson) || walkers.ApiHasBodyOfKind(api, spec.BodyString) {
		w.Imports.Module(g.Modules.ContentType)
	}
	w.Imports.Module(g.Modules.ServicesApi(api))
	w.Imports.Module(g.Modules.HttpErrors)
	w.Imports.Module(g.Modules.HttpErrorsModels)
	if walkers.ApiIsUsingModels(api) {
		w.Imports.Module(g.Modules.Models(api.InHttp.InVersion))
	}
	if operationHasParams(api) {
		w.Imports.Module(g.Modules.ParamsParser)
	}
	w.Imports.Module(g.Modules.Respond)

	w.EmptyLine()

	w.Line(`func %s(router *vestigo.Router, %s %s) {`, g.addRoutesMethodName(api), serviceInterfaceTypeVar(api), g.Modules.ServicesApi(api).Get(serviceInterfaceName))
	w.Indent()
	for _, operation := range api.Operations {
		url := g.getEndpointUrl(&operation)
		w.Line(`%s := log.Fields{"operationId": "%s.%s", "method": "%s", "url": "%s"}`, logFieldsName(&operation), operation.InApi.Name.Source, operation.Name.Source, casee.ToUpperCase(operation.Endpoint.Method), url)
		w.Line(`router.%s("%s", func(res http.ResponseWriter, req *http.Request) {`, casee.ToPascalCase(operation.Endpoint.Method), url)
		g.operation(w.Indented(), &operation)
		w.Line(`})`)
		if operation.HeaderParams != nil && len(operation.HeaderParams) > 0 {
			g.addSetCors(w, &operation)
		}
		w.EmptyLine()
	}
	w.Unindent()
	w.Line(`}`)

	return w.ToCodeFile()
}

func operationHasParams(api *spec.Api) bool {
	for _, operation := range api.Operations {
		for _, param := range operation.QueryParams {
			if &param != nil {
				return true
			}
		}
		for _, param := range operation.HeaderParams {
			if &param != nil {
				return true
			}
		}
		for _, param := range operation.Endpoint.UrlParams {
			if &param != nil {
				return true
			}
		}
	}
	return false
}

func (g *VestigoGenerator) getEndpointUrl(operation *spec.NamedOperation) string {
	url := operation.FullUrl()
	if operation.Endpoint.UrlParams != nil && len(operation.Endpoint.UrlParams) > 0 {
		for _, param := range operation.Endpoint.UrlParams {
			url = strings.Replace(url, spec.UrlParamStr(&param), fmt.Sprintf(":%s", param.Name.Source), -1)
		}
	}
	return url
}

func (g *VestigoGenerator) addSetCors(w *writer.Writer, operation *spec.NamedOperation) {
	w.Line(`router.SetCors("%s", &vestigo.CorsAccessControl{`, g.getEndpointUrl(operation))
	params := []string{}
	for _, param := range operation.HeaderParams {
		params = append(params, fmt.Sprintf(`"%s"`, param.Name.Source))
	}
	w.Line(`  AllowHeaders: []string{%s},`, strings.Join(params, ", "))
	w.Line(`})`)
}

func logFieldsName(operation *spec.NamedOperation) string {
	return fmt.Sprintf("log%s", operation.Name.PascalCase())
}

func (g *VestigoGenerator) parserParameterCall(isUrlParam bool, param *spec.NamedParam, paramsParserName string) string {
	paramNameSource := param.Name.Source
	if isUrlParam {
		paramNameSource = ":" + paramNameSource
	}
	parserParams := []string{fmt.Sprintf(`"%s"`, paramNameSource)}
	methodName, defaultParam := parserDefaultName(param)
	isEnum := param.Type.Definition.Info.Model != nil && param.Type.Definition.Info.Model.IsEnum()
	enumModel := param.Type.Definition.Info.Model
	if isEnum {
		parserParams = append(parserParams, fmt.Sprintf("%s.%s", types.VersionModelsPackage, g.Models.EnumValuesStrings(enumModel)))
	}
	if defaultParam != nil {
		parserParams = append(parserParams, *defaultParam)
	}
	call := fmt.Sprintf(`%s.%s(%s)`, paramsParserName, methodName, strings.Join(parserParams, ", "))
	if isEnum {
		call = fmt.Sprintf(`%s.%s(%s)`, types.VersionModelsPackage, enumModel.Name.PascalCase(), call)
	}
	return call
}

func (g *VestigoGenerator) headerParsing(w *writer.Writer, operation *spec.NamedOperation) {
	g.parametersParsing(w, operation, operation.HeaderParams, "header", "req.Header")
}

func (g *VestigoGenerator) queryParsing(w *writer.Writer, operation *spec.NamedOperation) {
	g.parametersParsing(w, operation, operation.QueryParams, "query", "req.URL.Query()")
}

func (g *VestigoGenerator) urlParamsParsing(w *writer.Writer, operation *spec.NamedOperation) {
	if operation.Endpoint.UrlParams != nil && len(operation.Endpoint.UrlParams) > 0 {
		w.Line(`urlParams := paramsparser.New(req.URL.Query(), false)`)
		for _, param := range operation.Endpoint.UrlParams {
			w.Line(`%s := %s`, param.Name.CamelCase(), g.parserParameterCall(true, &param, "urlParams"))
		}
		w.Line(`if len(urlParams.Errors) > 0 {`)
		g.respondNotFound(w.Indented(), operation, fmt.Sprintf(`"Failed to parse url parameters"`))
		w.Line(`}`)
	}
}

func (g *VestigoGenerator) parametersParsing(w *writer.Writer, operation *spec.NamedOperation, namedParams []spec.NamedParam, paramsParserName string, paramsValuesVar string) {
	if namedParams != nil && len(namedParams) > 0 {
		w.Line(`%s := paramsparser.New(%s, true)`, paramsParserName, paramsValuesVar)
		for _, param := range namedParams {
			w.Line(`%s := %s`, param.Name.CamelCase(), g.parserParameterCall(false, &param, paramsParserName))
		}

		w.Line(`if len(%s.Errors) > 0 {`, paramsParserName)
		g.respondBadRequest(w.Indented(), operation, paramsParserName, fmt.Sprintf(`"Failed to parse %s"`, paramsParserName), fmt.Sprintf(`httperrors.Convert(%s.Errors)`, paramsParserName))
		w.Line(`}`)
	}
}

func (g *VestigoGenerator) serviceCallAndResponseCheck(w *writer.Writer, operation *spec.NamedOperation, responseVar string) {
	singleEmptyResponse := len(operation.Responses) == 1 && operation.Responses[0].Type.Definition.IsEmpty()
	serviceCall := g.serviceCall(serviceInterfaceTypeVar(operation.InApi), operation)
	if singleEmptyResponse {
		w.Line(`err = %s`, serviceCall)
	} else {
		w.Line(`%s, err := %s`, responseVar, serviceCall)
	}

	w.Line(`if err != nil {`)
	g.respondInternalServerError(w.Indented(), operation, genFmtSprintf("Error returned from service implementation: %s", `err.Error()`))
	w.Line(`}`)

	if !singleEmptyResponse {
		w.Line(`if response == nil {`)
		g.respondInternalServerError(w.Indented(), operation, `"Service implementation returned nil"`)
		w.Line(`}`)
	}
}

func (g *VestigoGenerator) WriteResponse(w *writer.Writer, logFieldsName string, response *spec.Response, responseVar string) {
	if response.BodyIs(spec.BodyEmpty) {
		w.Line(respondEmpty(logFieldsName, `res`, spec.HttpStatusCode(response.Name)))
	}
	if response.BodyIs(spec.BodyString) {
		w.Line(respondText(logFieldsName, `res`, spec.HttpStatusCode(response.Name), `*`+responseVar))
	}
	if response.BodyIs(spec.BodyJson) {
		w.Line(respondJson(logFieldsName, `res`, spec.HttpStatusCode(response.Name), responseVar))
	}
}

func (g *VestigoGenerator) operation(w *writer.Writer, operation *spec.NamedOperation) {
	w.Line(`log.WithFields(%s).Info("Received request")`, logFieldsName(operation))
	w.Line(`var err error`)
	g.urlParamsParsing(w, operation)
	g.headerParsing(w, operation)
	g.queryParsing(w, operation)
	g.bodyParsing(w, operation)
	g.serviceCallAndResponseCheck(w, operation, `response`)
	g.response(w, operation, `response`)
}

func (g *VestigoGenerator) response(w *writer.Writer, operation *spec.NamedOperation, responseVar string) {
	if len(operation.Responses) == 1 {
		g.WriteResponse(w, logFieldsName(operation), &operation.Responses[0].Response, responseVar)
	} else {
		for _, response := range operation.Responses {
			responseVar := fmt.Sprintf("%s.%s", responseVar, response.Name.PascalCase())
			w.Line(`if %s != nil {`, responseVar)
			g.WriteResponse(w.Indented(), logFieldsName(operation), &response.Response, responseVar)
			w.Line(`  return`)
			w.Line(`}`)
		}
		g.respondInternalServerError(w, operation, `"Result from service implementation does not have anything in it"`)
	}
}

func (g *VestigoGenerator) bodyParsing(w *writer.Writer, operation *spec.NamedOperation) {
	if operation.BodyIs(spec.BodyString) {
		w.Line(`if !%s {`, callCheckContentType(logFieldsName(operation), `"text/plain"`, "req", "res"))
		w.Line(`  return`)
		w.Line(`}`)
		w.Line(`bodyData, err := ioutil.ReadAll(req.Body)`)
		w.Line(`if err != nil {`)
		g.respondBadRequest(w.Indented(), operation, "body", genFmtSprintf(`Reading request body failed: %s`, `err.Error()`), "nil")
		w.Line(`}`)
		w.Line(`body := string(bodyData)`)
	}
	if operation.BodyIs(spec.BodyJson) {
		w.Line(`if !%s {`, callCheckContentType(logFieldsName(operation), `"application/json"`, "req", "res"))
		w.Line(`  return`)
		w.Line(`}`)
		w.Line(`var body %s`, g.Types.GoType(&operation.Body.Type.Definition))
		w.Line(`err = json.NewDecoder(req.Body).Decode(&body)`)
		w.Line(`if err != nil {`)
		w.Line(`  var errors []errmodels.ValidationError = nil`)
		w.Line(`  if unmarshalError, ok := err.(*json.UnmarshalTypeError); ok {`)
		w.Line(`    message := fmt.Sprintf("Failed to parse JSON, field: PERCENT_s", unmarshalError.Field)`)
		w.Line(`    errors = []errmodels.ValidationError{{Path: unmarshalError.Field, Code: "parsing_failed", Message: &message}}`)
		w.Line(`  }`)
		g.respondBadRequest(w.Indented(), operation, "body", `"Failed to parse body"`, "errors")
		w.Line(`}`)
	}
}

func (g *VestigoGenerator) serviceCall(serviceVar string, operation *spec.NamedOperation) string {
	params := []string{}
	if operation.BodyIs(spec.BodyString) {
		params = append(params, "body")
	}
	if operation.BodyIs(spec.BodyJson) {
		params = append(params, "&body")
	}
	for _, param := range operation.QueryParams {
		params = append(params, param.Name.CamelCase())
	}
	for _, param := range operation.HeaderParams {
		params = append(params, param.Name.CamelCase())
	}
	for _, param := range operation.Endpoint.UrlParams {
		params = append(params, param.Name.CamelCase())
	}

	return fmt.Sprintf(`%s.%s(%s)`, serviceVar, operation.Name.PascalCase(), strings.Join(params, ", "))
}

func (g *VestigoGenerator) addRoutesMethodName(api *spec.Api) string {
	return fmt.Sprintf(`Add%sRoutes`, api.Name.PascalCase())
}

func genFmtSprintf(format string, args ...string) string {
	if len(args) > 0 {
		return fmt.Sprintf(`fmt.Sprintf("%s", %s)`, format, strings.Join(args, ", "))
	} else {
		return format
	}
}

func serviceInterfaceTypeVar(api *spec.Api) string {
	return fmt.Sprintf(`%sService`, api.Name.Source)
}

func (g *VestigoGenerator) RootRouting(specification *spec.Spec) *generator.CodeFile {
	w := writer.New(g.Modules.Root, "spec.go")

	w.Imports.Add("github.com/husobee/vestigo")
	for _, version := range specification.Versions {
		w.Imports.ModuleAliased(g.Modules.Routing(&version).Aliased(routingPackageAlias(&version)))
		for _, api := range version.Http.Apis {
			w.Imports.ModuleAliased(g.Modules.ServicesApi(&api).Aliased(apiPackageAlias(&api)))
		}
	}

	w.EmptyLine()
	routesParams := []string{}
	for _, version := range specification.Versions {
		for _, api := range version.Http.Apis {
			apiModule := g.Modules.ServicesApi(&api).Aliased(apiPackageAlias(&api))
			routesParams = append(routesParams, fmt.Sprintf(`%s %s`, serviceApiNameVersioned(&api), apiModule.Get(serviceInterfaceName)))
		}
	}
	w.Line(`func AddRoutes(router *vestigo.Router, %s) {`, strings.Join(routesParams, ", "))
	for _, version := range specification.Versions {
		routingModule := g.Modules.Routing(&version).Aliased(routingPackageAlias(&version))
		for _, api := range version.Http.Apis {
			w.Line(`  %s(router, %s)`, routingModule.Get(g.addRoutesMethodName(&api)), serviceApiNameVersioned(&api))
		}
	}
	w.Line(`}`)

	return w.ToCodeFile()
}

func routingPackageAlias(version *spec.Version) string {
	if version.Name.Source != "" {
		return fmt.Sprintf(`%s`, version.Name.FlatCase())
	} else {
		return fmt.Sprintf(`root`)
	}
}

func apiPackageAlias(api *spec.Api) string {
	version := api.InHttp.InVersion.Name
	if version.Source != "" {
		return api.Name.CamelCase() + version.PascalCase()
	}
	return api.Name.CamelCase()
}

func serviceApiNameVersioned(api *spec.Api) string {
	return fmt.Sprintf(`%sService%s`, api.Name.Source, api.InHttp.InVersion.Name.PascalCase())
}

func (g *VestigoGenerator) CheckContentType() *generator.CodeFile {
	w := writer.New(g.Modules.ContentType, `check.go`)
	w.Template(
		map[string]string{
			`ErrorsPackage`:       g.Modules.HttpErrors.Package,
			`ErrorsModelsPackage`: g.Modules.HttpErrorsModels.Package,
		}, `
import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"[[.ErrorsPackage]]"
	"[[.ErrorsModelsPackage]]"
)

func Check(logFields log.Fields, expectedContentType string, req *http.Request, res http.ResponseWriter) bool {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, expectedContentType) {
		message := fmt.Sprintf("Expected Content-Type header: '%s' was not provided, found: '%s'", expectedContentType, contentType)
		httperrors.RespondBadRequest(logFields, res, &errmodels.BadRequestError{Location: "header", Message: "Failed to parse header", Errors: []errmodels.ValidationError{{Path: "Content-Type", Code: "missing", Message: &message}}})
		return false
	}
	return true
}
`)
	return w.ToCodeFile()
}

func (g *VestigoGenerator) HttpErrors(responses *spec.Responses) []generator.CodeFile {
	files := []generator.CodeFile{}

	files = append(files, *g.errorsModelsConverter())
	files = append(files, *g.ErrorResponses(responses))

	return files
}

func (g *VestigoGenerator) errorsModelsConverter() *generator.CodeFile {
	w := writer.New(g.Modules.HttpErrors, `converter.go`)
	w.Template(
		map[string]string{
			`ErrorsModelsPackage`: g.Modules.HttpErrorsModels.Package,
			`ParamsParserModule`:  g.Modules.ParamsParser.Package,
		}, `
import (
	"[[.ErrorsModelsPackage]]"
	"[[.ParamsParserModule]]"
)

func Convert(parsingErrors []paramsparser.ParsingError) []errmodels.ValidationError {
	var validationErrors []errmodels.ValidationError

	for _, parsingError := range parsingErrors {
		validationError := errmodels.ValidationError(parsingError)
		validationErrors = append(validationErrors, validationError)
	}

	return validationErrors
}
`)
	return w.ToCodeFile()
}
