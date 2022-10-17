package service

import (
	"fmt"
	"github.com/specgen-io/specgen-golang/v2/imports"
	"github.com/specgen-io/specgen-golang/v2/module"
	"github.com/specgen-io/specgen-golang/v2/writer"

	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
)

func (g *VestigoGenerator) generateErrors(converterModule, errorsModelsModule, respondModule module.Module, responses *spec.Responses) *generator.CodeFile {
	w := writer.New(converterModule, "responses.go")

	imports := imports.New()
	imports.AddAliased("github.com/sirupsen/logrus", "log")
	imports.Add("net/http")
	imports.Module(errorsModelsModule)
	imports.Module(respondModule)
	imports.Write(w)

	w.EmptyLine()
	badRequest := responses.GetByStatusName(spec.HttpStatusBadRequest)
	w.Line(`func RespondBadRequest(logFields log.Fields, res http.ResponseWriter, error *%s) {`, g.Types.GoType(&badRequest.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	g.generateResponseWriting(w.Indented(), `logFields`, badRequest, `error`)
	w.Line(`}`)

	w.EmptyLine()
	notFound := responses.GetByStatusName(spec.HttpStatusNotFound)
	w.Line(`func RespondNotFound(logFields log.Fields, res http.ResponseWriter, error *%s) {`, g.Types.GoType(&notFound.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	g.generateResponseWriting(w.Indented(), `logFields`, notFound, `error`)
	w.Line(`}`)

	w.EmptyLine()
	internalServerError := responses.GetByStatusName(spec.HttpStatusInternalServerError)
	w.Line(`func RespondInternalServerError(logFields log.Fields, res http.ResponseWriter, error *%s) {`, g.Types.GoType(&internalServerError.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	g.generateResponseWriting(w.Indented(), `logFields`, internalServerError, `error`)
	w.Line(`}`)

	return w.ToCodeFile()
}

func callCheckContentType(logFieldsVar, expectedContentType, requestVar, responseVar string) string {
	return fmt.Sprintf(`contenttype.Check(%s, %s, %s, %s)`, logFieldsVar, expectedContentType, requestVar, responseVar)
}

func (g *VestigoGenerator) respondNotFound(w generator.Writer, operation *spec.NamedOperation, message string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	badRequest := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusNotFound)
	errorMessage := fmt.Sprintf(`%s{Message: %s}`, g.Types.GoType(&badRequest.Type.Definition), message)
	w.Line(`httperrors.RespondNotFound(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}

func (g *VestigoGenerator) respondBadRequest(w generator.Writer, operation *spec.NamedOperation, location string, message string, params string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	badRequest := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusBadRequest)
	errorMessage := fmt.Sprintf(`%s{Location: "%s", Message: %s, Errors: %s}`, g.Types.GoType(&badRequest.Type.Definition), location, message, params)
	w.Line(`httperrors.RespondBadRequest(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}

func (g *VestigoGenerator) respondInternalServerError(w generator.Writer, operation *spec.NamedOperation, message string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	internalServerError := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusInternalServerError)
	errorMessage := fmt.Sprintf(`%s{Message: %s}`, g.Types.GoType(&internalServerError.Type.Definition), message)
	w.Line(`httperrors.RespondInternalServerError(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}
